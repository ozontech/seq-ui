package types

import "time"

type SaveAsyncSearchRequest struct {
	SearchID  string
	OwnerID   int64
	ExpiresAt time.Time
	Meta      string
}

type AsyncSearchInfo struct {
	SearchID  string
	OwnerID   int64
	OwnerName string
	Meta      string
	CreatedAt time.Time
	ExpiresAt time.Time
}

type GetAsyncSearchesListRequest struct {
	Limit  int32
	Offset int32
	Owner  *string
}
