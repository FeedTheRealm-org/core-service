package dtos

import "time"

type AppName string

const (
	AppNameWorldEditor AppName = "ftr_world_editor"
	AppNameGame        AppName = "ftr_game"
)

func (a AppName) String() string {
	return string(a)
}

func (a AppName) Valid() bool {
	switch a {
	case AppNameWorldEditor, AppNameGame:
		return true
	default:
		return false
	}
}

type OSName string

const (
	OSNameLinux   OSName = "linux"
	OSNameWindows OSName = "windows"
)

func (o OSName) String() string {
	return string(o)
}

func (o OSName) Valid() bool {
	switch o {
	case OSNameLinux, OSNameWindows:
		return true
	default:
		return false
	}
}

type ExportZipResponse struct {
	AppName   string    `json:"app_name"`
	Version   string    `json:"version"`
	OS        string    `json:"os"`
	Path      string    `json:"path"`
	IsLatest  bool      `json:"is_latest"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ExportZipPathResponse struct {
	Path string `json:"path"`
}

type ExportZipSetLatestRequest struct {
	AppName string `json:"app_name"`
	Version string `json:"version"`
	OS      string `json:"os"`
}
