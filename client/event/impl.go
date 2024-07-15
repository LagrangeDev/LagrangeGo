package event

type INotifyEvent interface {
	From() uint32
	Content() string
}
