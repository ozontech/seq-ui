package sessionstore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/ozontech/seq-ui/internal/app/config/v2"
	"github.com/ozontech/seq-ui/internal/app/types"
	"github.com/ozontech/seq-ui/internal/pkg/cache"
	"github.com/ozontech/seq-ui/logger"
)

type cacheSessionStore struct {
	cache     cache.Store
	lifetime  time.Duration
	keyPrefix string
}

const (
	activeSessionIDKey = "active_session_id"
	exportsKeysPattern = "job#*#export#*"
)

func NewCacheSessionStore(ctx context.Context, cfg *config.SessionStore, env string) (SessionStore, error) {
	c, ok := cache.StoreByID(cfg.RedisCacheID)
	if !ok {
		return nil, fmt.Errorf("cache %q not found", cfg.RedisCacheID)
	}

	var keyPrefix string
	if env != "" {
		keyPrefix = env + "_"
	}

	store := &cacheSessionStore{
		cache:     c,
		keyPrefix: keyPrefix,
		lifetime:  cfg.Lifetime,
	}

	store.unlock(ctx)

	return store, nil
}

func (s *cacheSessionStore) StartExport(ctx context.Context, sessionID string, info types.ExportInfo) error {
	err := s.Lock(ctx, sessionID)
	if err != nil {
		return err
	}

	return s.set(ctx, sessionID, info)
}

func (s *cacheSessionStore) CheckExport(ctx context.Context, sessionID string) (types.ExportInfo, error) {
	return s.get(ctx, sessionID)
}

func (s *cacheSessionStore) GetCurActiveExport(ctx context.Context) (string, error) {
	activeSessionID, err := s.cache.Get(ctx, s.keyPrefix+activeSessionIDKey)
	if err == nil {
		return activeSessionID, nil
	}

	if errors.Is(err, cache.ErrNotFound) {
		return "", ErrNoActiveExport
	}

	return "", fmt.Errorf("get: %w", err)
}

func (s *cacheSessionStore) CancelExport(ctx context.Context, sessionID string) error {
	s.unlock(ctx)

	return s.setExportStatus(ctx, sessionID, types.ExportStatusCancel, "")
}

func (s *cacheSessionStore) FailExport(ctx context.Context, sessionID string, errMsg string) error {
	s.unlock(ctx)

	return s.setExportStatus(ctx, sessionID, types.ExportStatusFail, errMsg)
}

func (s *cacheSessionStore) FinishExport(ctx context.Context, sessionID string) error {
	s.unlock(ctx)

	return s.setExportStatus(ctx, sessionID, types.ExportStatusFinish, "")
}

func (s *cacheSessionStore) Lock(ctx context.Context, sessionID string) error {
	ok, err := s.cache.SetNX(ctx, s.keyPrefix+activeSessionIDKey, sessionID, s.lifetime)
	if err != nil {
		return fmt.Errorf("set: %w", err)
	}

	if !ok {
		activeSessionID, err := s.cache.Get(ctx, s.keyPrefix+activeSessionIDKey)
		if err != nil {
			return fmt.Errorf("get active session id: %w", err)
		}

		return fmt.Errorf("export with id '%s' already started (see redis)", activeSessionID)
	}

	return nil
}

func (s *cacheSessionStore) GetAllExports(ctx context.Context) ([]types.ExportInfo, error) {
	keys, err := s.cache.KeysByPattern(ctx, s.keyPrefix+exportsKeysPattern)
	if err != nil {
		return nil, fmt.Errorf("get user exports: %w", err)
	}

	// special case: mget can't take 0 keys
	if len(keys) == 0 {
		return nil, nil
	}

	values, err := s.cache.MGet(ctx, keys...)
	if err != nil {
		return nil, fmt.Errorf("get values: %w", err)
	}

	result := make([]types.ExportInfo, 0, len(values))
	for i, value := range values {
		valueStr, ok := value.(string)
		if !ok {
			return nil, errors.New("type mismatch; expected string")
		}

		var localInfo exportInfo
		err = json.Unmarshal([]byte(valueStr), &localInfo)
		if err != nil {
			logger.Error("can't unmarshal; problem can be caused by changing format", zap.Error(err))
			continue
		}

		var globalInfo types.ExportInfo
		globalInfo, err = localToGlobal(localInfo, strings.TrimPrefix(keys[i], s.keyPrefix))
		if err != nil {
			logger.Error("can't convert export info", zap.Error(err))
			continue
		}

		result = append(result, globalInfo)
	}

	return result, nil
}

func (s *cacheSessionStore) ConfirmPart(ctx context.Context, sessionID string, partID int, partSize types.Size) error {
	info, err := s.get(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("get: %w", err)
	}

	if partID < 0 || partID >= len(info.PartIsUploaded) {
		return fmt.Errorf("index out of range; index = %d; parts count = %d", partID, len(info.PartIsUploaded))
	}

	if info.PartIsUploaded[partID] {
		return fmt.Errorf("part is already confirmed; index = %d", partID)
	}

	info.PartIsUploaded[partID] = true
	info.TotalSize.Unpacked += partSize.Unpacked
	info.TotalSize.Packed += partSize.Packed

	return s.set(ctx, sessionID, info)
}

func (s *cacheSessionStore) setExportStatus(
	ctx context.Context,
	sessionID string,
	status types.ExportStatus,
	errMsg string,
) error {
	result, err := s.get(ctx, sessionID)
	if err != nil {
		return err
	}

	switch result.Status {
	case types.ExportStatusCancel:
		return errors.New("export already canceled")
	case types.ExportStatusFail:
		return errors.New("export already failed")
	case types.ExportStatusFinish:
		return errors.New("export already finished")
	default:
		result.Status = status
		result.Error = errMsg
		result.FinishedAt = time.Now()
		return s.set(ctx, sessionID, result)
	}
}

func (s *cacheSessionStore) set(ctx context.Context, sessionID string, info types.ExportInfo) error {
	data, err := json.Marshal(globalToLocal(info))
	if err != nil {
		return fmt.Errorf("pack export info: %w", err)
	}

	if err := s.cache.SetWithTTL(ctx, s.keyPrefix+sessionID, string(data), s.lifetime); err != nil {
		return fmt.Errorf("set: %w", err)
	}

	return nil
}

func (s *cacheSessionStore) get(ctx context.Context, sessionID string) (types.ExportInfo, error) {
	str, err := s.cache.Get(ctx, s.keyPrefix+sessionID)
	if err != nil {
		if errors.Is(err, cache.ErrNotFound) {
			return types.ExportInfo{}, types.ErrNotFound
		}
		return types.ExportInfo{}, fmt.Errorf("get: %w", err)
	}

	var info exportInfo
	err = json.Unmarshal([]byte(str), &info)
	if err != nil {
		return types.ExportInfo{}, fmt.Errorf("unpack export info: %w", err)
	}

	result, err := localToGlobal(info, sessionID)
	if err != nil {
		return types.ExportInfo{}, fmt.Errorf("local to global: %w", err)
	}

	return result, nil
}

func (s *cacheSessionStore) unlock(ctx context.Context) {
	s.cache.Del(ctx, s.keyPrefix+activeSessionIDKey)
}

func strToBoolArray(s string) ([]bool, error) {
	a := make([]bool, len(s))
	for i, c := range s {
		switch c {
		case '0':
			a[i] = false
		case '1':
			a[i] = true
		default:
			return nil, fmt.Errorf("invalid char: %c", c)
		}
	}

	return a, nil
}

func boolArrayToStr(a []bool) string {
	var b strings.Builder
	for _, val := range a {
		if val {
			b.WriteByte('1')
		} else {
			b.WriteByte('0')
		}
	}

	return b.String()
}

func getProgress(partConfirmed []bool) float64 {
	count := 0
	for i := range partConfirmed {
		if partConfirmed[i] {
			count++
		}
	}

	return float64(count) / float64(len(partConfirmed))
}

func extractUserFromSessionID(sessionID string) (string, error) {
	tokens := strings.Split(sessionID, "#")
	if len(tokens) != 4 {
		return "", fmt.Errorf("wrong tokens count: %d (expected 4)", len(tokens))
	}

	return tokens[1], nil
}

type exportInfo struct {
	Status types.ExportStatus `json:"status"`

	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	StartedAt  time.Time `json:"started_at"`
	FinishedAt time.Time `json:"finished_at"`

	PartIsUploaded string `json:"part_is_uploaded"`

	TotalSize totalSize `json:"total_size"`

	FileStorePathPrefix string `json:"file_store_path_prefix"`

	From       time.Time     `json:"from"`
	To         time.Time     `json:"to"`
	Query      string        `json:"query"`
	Window     time.Duration `json:"window"`
	BatchSize  uint64        `json:"batch_size"`
	PartLength time.Duration `json:"part_length"`

	Error string `json:"error"`
}

type totalSize struct {
	Unpacked int `json:"unpacked"`
	Packed   int `json:"packed"`
}

func globalToLocal(info types.ExportInfo) exportInfo {
	return exportInfo{
		Status: info.Status,
		Error:  info.Error,

		CreatedAt:  info.CreatedAt,
		UpdatedAt:  info.UpdatedAt,
		StartedAt:  info.StartedAt,
		FinishedAt: info.FinishedAt,

		PartIsUploaded: boolArrayToStr(info.PartIsUploaded),

		TotalSize: totalSize{
			Unpacked: info.TotalSize.Unpacked,
			Packed:   info.TotalSize.Packed,
		},

		FileStorePathPrefix: info.FileStorePathPrefix,

		From:       info.From,
		To:         info.To,
		Query:      info.Query,
		Window:     info.Window,
		BatchSize:  info.BatchSize,
		PartLength: info.PartLength,
	}
}

func localToGlobal(info exportInfo, sessionID string) (types.ExportInfo, error) {
	user, err := extractUserFromSessionID(sessionID)
	if err != nil {
		return types.ExportInfo{}, fmt.Errorf("extract user from session id: %w", err)
	}

	partIsUploaded, err := strToBoolArray(info.PartIsUploaded)
	if err != nil {
		return types.ExportInfo{}, fmt.Errorf("str to bool array: %w", err)
	}

	return types.ExportInfo{
		ID:     sessionID,
		UserID: user,

		Status:   info.Status,
		Error:    info.Error,
		Progress: getProgress(partIsUploaded),

		CreatedAt:  info.CreatedAt,
		UpdatedAt:  info.UpdatedAt,
		StartedAt:  info.StartedAt,
		FinishedAt: info.FinishedAt,

		PartIsUploaded: partIsUploaded,

		TotalSize: types.Size{
			Unpacked: info.TotalSize.Unpacked,
			Packed:   info.TotalSize.Packed,
		},

		FileStorePathPrefix: info.FileStorePathPrefix,

		From:       info.From,
		To:         info.To,
		Query:      info.Query,
		Window:     info.Window,
		BatchSize:  info.BatchSize,
		PartLength: info.PartLength,
	}, nil
}
