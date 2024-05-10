package client

// from https://github.com/Mrs4s/MiraiGo/blob/master/client/global.go

import (
	"net"
	"time"

	"github.com/pkg/errors"
)

func qualityTest(addr string) (int64, error) {
	// see QualityTestManager
	start := time.Now()
	conn, err := net.DialTimeout("tcp", addr, time.Second*5)
	if err != nil {
		return 0, errors.Wrap(err, "failed to connect to server during quality test")
	}
	_ = conn.Close()
	return time.Since(start).Milliseconds(), nil
}
