package model

type RapidFile struct {
	ID   string `json:"uuid"` // uuid
	Name string `json:"name"` // filename without the extension
	Ext  string `json:"ext"`  // extension of file
	Path string `json:"path"` // path to file + filename
	Size int64  `json:"size"` // size in bytes
}
