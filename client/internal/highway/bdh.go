package highway

// from https://github.com/Mrs4s/MiraiGo/tree/master/client/internal/highway/bdh.go

import (
	"crypto/md5"
	"io"
	"strconv"
	"sync"
	"sync/atomic"

	ftea "github.com/fumiama/gofastTEA"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"

	"github.com/LagrangeDev/LagrangeGo/client/packets/pb/service/highway"
	"github.com/LagrangeDev/LagrangeGo/internal/proto"
	"github.com/LagrangeDev/LagrangeGo/utils/binary"
)

const BlockSize = 256 * 1024

type Transaction struct {
	CommandID uint32
	Body      io.Reader
	Sum       []byte // md5 sum of body
	Size      uint64 // body size
	Ticket    []byte
	LoginSig  []byte
	Ext       []byte
	Encrypt   bool
}

func (trans *Transaction) encrypt(key []byte) error {
	if !trans.Encrypt {
		return nil
	}
	if len(key) == 0 {
		return errors.New("session key not found. maybe miss some packet?")
	}
	trans.Ext = ftea.NewTeaCipher(key).Encrypt(trans.Ext)
	return nil
}

func (trans *Transaction) Build(s *Session, offset uint64, length uint32, md5hash []byte) *highway.ReqDataHighwayHead {
	return &highway.ReqDataHighwayHead{
		MsgBaseHead: &highway.DataHighwayHead{
			Version:    1,
			Uin:        proto.Some(strconv.Itoa(int(*s.Uin))),
			Command:    proto.Some(_REQ_CMD_DATA),
			Seq:        proto.Some(s.NextSeq()),
			RetryTimes: proto.Some(uint32(0)),
			AppId:      s.SubAppID,
			DataFlag:   16,
			CommandId:  trans.CommandID,
			// LocaleId:  2052,
		},
		MsgSegHead: &highway.SegHead{
			ServiceId:     proto.Some(uint32(0)),
			Filesize:      trans.Size,
			DataOffset:    proto.Some(offset),
			DataLength:    length,
			RetCode:       proto.Some(uint32(0)),
			ServiceTicket: trans.Ticket,
			Md5:           md5hash,
			FileMd5:       trans.Sum,
			CacheAddr:     proto.Some(uint32(0)),
			CachePort:     proto.Some(uint32(0)),
		},
		BytesReqExtendInfo: trans.Ext,
		MsgLoginSigHead: &highway.LoginSigHead{
			Uint32LoginSigType: 8,
			BytesLoginSig:      trans.LoginSig,
			AppId:              s.AppID,
		},
	}
}

func (s *Session) uploadSingle(trans *Transaction) ([]byte, error) {
	pc, err := s.selectConn()
	if err != nil {
		return nil, err
	}
	defer s.putIdleConn(pc)

	reader := binary.NewNetworkReader(pc.conn)
	var rspExt []byte
	offset := 0
	chunk := make([]byte, BlockSize)
	for {
		chunk = chunk[:cap(chunk)]
		rl, err := io.ReadFull(trans.Body, chunk)
		if rl == 0 {
			break
		}
		if errors.Is(err, io.ErrUnexpectedEOF) {
			chunk = chunk[:rl]
		}
		ch := md5.Sum(chunk)
		head, _ := proto.Marshal(trans.Build(s, uint64(offset), uint32(rl), ch[:]))
		offset += rl
		buffers := Frame(head, chunk)
		_, err = buffers.WriteTo(pc.conn)
		if err != nil {
			return nil, errors.Wrap(err, "write conn error")
		}
		rspHead, err := readResponse(reader)
		if err != nil {
			return nil, errors.Wrap(err, "highway upload error")
		}
		if rspHead.ErrorCode != 0 {
			return nil, errors.Errorf("upload failed: %d", rspHead.ErrorCode)
		}
		if rspHead.BytesRspExtendInfo != nil {
			rspExt = rspHead.BytesRspExtendInfo
		}
		if rspHead.MsgSegHead != nil && rspHead.MsgSegHead.ServiceTicket != nil {
			trans.Ticket = rspHead.MsgSegHead.ServiceTicket
		}
	}
	return rspExt, nil
}

func (s *Session) Upload(trans *Transaction) ([]byte, error) {
	// encrypt ext data
	if err := trans.encrypt(s.SessionKey); err != nil {
		return nil, err
	}

	const maxThreadCount = 4
	threadCount := int(trans.Size) / (6 * BlockSize) // 1 thread upload 1.5 MB
	if threadCount > maxThreadCount {
		threadCount = maxThreadCount
	}
	if threadCount < 2 {
		// single thread upload
		return s.uploadSingle(trans)
	}

	// pick a address
	// TODO: pick smarter
	pc, err := s.selectConn()
	if err != nil {
		return nil, err
	}
	addr := pc.addr
	s.putIdleConn(pc)

	var (
		rspExt          []byte
		completedThread uint32
		cond            = sync.NewCond(&sync.Mutex{})
		offset          = uint64(0)
		count           = (trans.Size + BlockSize - 1) / BlockSize
		id              = 0
	)
	doUpload := func() error {
		// send signal complete uploading
		defer func() {
			atomic.AddUint32(&completedThread, 1)
			cond.Signal()
		}()

		// todo: get from pool?
		pc, err := s.connect(addr)
		if err != nil {
			return err
		}
		defer s.putIdleConn(pc)

		reader := binary.NewNetworkReader(pc.conn)
		chunk := make([]byte, BlockSize)
		for {
			cond.L.Lock() // lock protect reading
			off := offset
			offset += BlockSize
			id++
			last := uint64(id) == count
			if last { // last
				for atomic.LoadUint32(&completedThread) != uint32(threadCount-1) {
					cond.Wait()
				}
			} else if uint64(id) > count {
				cond.L.Unlock()
				break
			}
			chunk = chunk[:BlockSize]
			n, err := io.ReadFull(trans.Body, chunk)
			cond.L.Unlock()

			if n == 0 {
				break
			}
			if errors.Is(err, io.ErrUnexpectedEOF) {
				chunk = chunk[:n]
			}
			ch := md5.Sum(chunk)
			head, _ := proto.Marshal(trans.Build(s, off, uint32(n), ch[:]))
			buffers := Frame(head, chunk)
			_, err = buffers.WriteTo(pc.conn)
			if err != nil {
				return errors.Wrap(err, "write conn error")
			}
			rspHead, err := readResponse(reader)
			if err != nil {
				return errors.Wrap(err, "highway upload error")
			}
			if rspHead.ErrorCode != 0 {
				return errors.Errorf("upload failed: %d", rspHead.ErrorCode)
			}
			if last && rspHead.BytesRspExtendInfo != nil {
				rspExt = rspHead.BytesRspExtendInfo
			}
		}
		return nil
	}

	group := errgroup.Group{}
	for i := 0; i < threadCount; i++ {
		group.Go(doUpload)
	}
	return rspExt, group.Wait()
}
