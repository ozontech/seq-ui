package filestore

import (
	"context"
	"io"
)

type FileStore interface {
	PutObject(ctx context.Context, objectName string, reader io.Reader) error
}
