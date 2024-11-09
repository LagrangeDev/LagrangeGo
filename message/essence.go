package message

type GroupEssenceMessage struct {
	OperatorUin  uint32
	OperatorUID  string
	OperatorTime uint64
	CanRemove    bool
	Message      *GroupMessage
}
