package entity

type GroupFile struct {
	GroupUin      uint32
	FileId        string
	FileName      string
	BusId         uint32
	FileSize      uint64
	UploadTime    uint32
	DeadTime      uint32
	ModifyTime    uint32
	DownloadTimes uint32
	Uploader      uint32
	UploaderName  string
}

type GroupFolder struct {
	GroupUin       uint32
	FolderId       string
	FolderName     string
	CreateTime     uint32
	Creator        uint32
	CreatorName    string
	TotalFileCount uint32
}
