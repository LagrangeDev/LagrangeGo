package entity

type GroupFileSystemInfo struct {
	GroupUin   uint32 `json:"group_id"`
	FileCount  uint32 `json:"file_count"`
	LimitCount uint32 `json:"limit_count"`
	UsedSpace  uint64 `json:"used_space"`
	TotalSpace uint64 `json:"total_space"`
}

type GroupFile struct {
	GroupUin      uint32 `json:"group_id"`
	FileID        string `json:"file_id"`
	FileName      string `json:"file_name"`
	BusID         uint32 `json:"busid"`
	FileSize      uint64 `json:"file_size"`
	UploadTime    uint32 `json:"upload_time"`
	DeadTime      uint32 `json:"dead_time"`
	ModifyTime    uint32 `json:"modify_time"`
	DownloadTimes uint32 `json:"download_times"`
	Uploader      uint32 `json:"uploader"`
	UploaderName  string `json:"uploader_name"`
}

type GroupFolder struct {
	GroupUin       uint32 `json:"group_id"`
	FolderID       string `json:"folder_id"`
	FolderName     string `json:"folder_name"`
	CreateTime     uint32 `json:"create_time"`
	Creator        uint32 `json:"creator"`
	CreatorName    string `json:"creator_name"`
	TotalFileCount uint32 `json:"total_file_count"`
}
