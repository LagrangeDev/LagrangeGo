package message

type GroupImageElement struct {
	ImageId   string
	FileId    int64
	ImageType int32
	Size      uint32
	Width     int32
	Height    int32
	Md5       []byte
	Url       string

	// EffectID show pic effect id.
	EffectID int32
	Flash    bool
}

type FriendImageElement struct {
	ImageId string
	Md5     []byte
	Size    uint32
	Width   int32
	Height  int32
	Url     string

	Flash bool
}
