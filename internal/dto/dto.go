package dto

// dto to map form fields when adding download
type AddDownloadDto struct {
	Url      string `json:"url"`
	Dir      string `json:"dir"`
	Filename string `json:"filename"`
}

type FileResponse struct {
	Id       int64  `json:"id"`
	Filename string `json:"filename"`
}

// JSON for basic current dir info
type CurDirInfo struct {
	Path      string `json:"path"`
	FreeSpace string `json:"freeSpace"`
}

type LoginDto struct {
	Password string `json:"password"`
}

type MsgResponse struct {
	Msg string `json:"msg"`
}
