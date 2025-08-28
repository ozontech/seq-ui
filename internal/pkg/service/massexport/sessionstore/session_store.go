package sessionstore

import (
	"context"
	"errors"

	"github.com/ozontech/seq-ui/internal/app/types"
)

var ErrNoActiveExport = errors.New("no active export now")

type SessionStore interface {
	StartExport(ctx context.Context, sessionID string, info types.ExportInfo) error
	CheckExport(ctx context.Context, sessionID string) (types.ExportInfo, error)
	CancelExport(ctx context.Context, sessionID string) error
	FailExport(ctx context.Context, sessionID string, errMsg string) error
	FinishExport(ctx context.Context, sessionID string) error
	ConfirmPart(ctx context.Context, sessionID string, partID int, partSize types.Size) error
	GetCurActiveExport(ctx context.Context) (string, error)
	Lock(ctx context.Context, sessionID string) error
	GetAllExports(ctx context.Context) ([]types.ExportInfo, error)
}
