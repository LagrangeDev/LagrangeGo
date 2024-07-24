package entity

type EventState uint32

const (
	NoNeed EventState = iota
	Unprocessed
	Processed
)

type (
	Group struct {
		GroupUin    uint32
		GroupName   string
		MemberCount uint32
		MaxMember   uint32
		Avatar      string
	}

	GroupJoinRequest struct {
		GroupUin    uint32
		InvitorUin  uint32
		InvitorUid  string
		TargetUin   uint32
		TargetUid   string
		OperatorUin uint32
		OperatorUid string
		Sequence    uint64
		State       EventState // 0不需要处理 1未处理 2已处理
		EventType   uint32     // 1申请加群 3设置管理员 16取消管理员
		Comment     string
	}
)

func (r *GroupJoinRequest) Checked() bool {
	return r.State == Processed || r.State == Unprocessed
}
