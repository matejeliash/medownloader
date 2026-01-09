package dto

// dto to map internal downloaded item to json
type DownloadItemDto struct {
	Id         int64  `json:"id"`
	Url        string `json:"url"`
	Filename   string `json:"filename"`
	Filepath   string `json:"filepath"`
	Active     bool   `json:"active"`
	Completed  bool   `json:"completed"`
	Downloaded int64  `json:"downloaded"`
	Size       int64  `json:"size"`

	Err string `json:"err"`
}
