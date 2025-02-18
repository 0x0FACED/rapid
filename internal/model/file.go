package model

type File struct {
	ID   string `json:"uuid"` // uuid
	Name string `json:"name"` // filename
	Path string `json:"path"` // path to file + filename
	Size int64  `json:"size"` // size in bytes
}
