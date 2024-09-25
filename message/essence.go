package message

type GroupEssenceMessage struct {
	OperatorUin  uint32
	OperatorUid  string
	OperatorTime uint64
	CanRemove    bool
	Message      *GroupMessage
}
