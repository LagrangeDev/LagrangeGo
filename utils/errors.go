package utils

import "errors"

var (
	GrpSendFailed = errors.New("group message send failed")
	PrvSendFailed = errors.New("private message send failed")
)
