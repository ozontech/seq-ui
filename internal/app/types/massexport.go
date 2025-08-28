package types

import (
	"fmt"
	"time"
)

type StartExportRequest struct {
	Query  string
	From   time.Time
	To     time.Time
	Window time.Duration
	Name   string
}

type StartExportResponse struct {
	SessionID string
}

type ExportStatus int

const (
	ExportStatusUnspecified ExportStatus = iota
	ExportStatusStart
	ExportStatusCancel
	ExportStatusFail
	ExportStatusFinish
)

func (s ExportStatus) String() string {
	switch s {
	case ExportStatusUnspecified:
		return "unspecified"
	case ExportStatusStart:
		return "start"
	case ExportStatusCancel:
		return "cancel"
	case ExportStatusFail:
		return "fail"
	case ExportStatusFinish:
		return "finish"
	default:
		panic(fmt.Sprintf("unknown export status: %d", s))
	}
}

type ExportInfo struct {
	ID       string
	UserID   string
	Status   ExportStatus
	Error    string
	Progress float64

	CreatedAt  time.Time
	UpdatedAt  time.Time
	StartedAt  time.Time
	FinishedAt time.Time

	Links          string
	PartIsUploaded []bool

	TotalSize Size

	FileStorePathPrefix string

	From       time.Time
	To         time.Time
	Query      string
	Window     time.Duration
	BatchSize  uint64
	PartLength time.Duration
}

type Size struct {
	Unpacked int
	Packed   int
}
