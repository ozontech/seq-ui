package massexport

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/ozontech/seq-ui/internal/app/config"
	"github.com/ozontech/seq-ui/internal/app/types"
	"github.com/ozontech/seq-ui/internal/pkg/client/seqdb"
	"github.com/ozontech/seq-ui/internal/pkg/service/massexport/filestore"
	"github.com/ozontech/seq-ui/internal/pkg/service/massexport/sessionstore"
	"github.com/ozontech/seq-ui/logger"
)

type Service interface {
	StartExport(ctx context.Context, req types.StartExportRequest) (types.StartExportResponse, error)
	CheckExport(ctx context.Context, sessionID string) (types.ExportInfo, error)
	CancelExport(ctx context.Context, sessionID string) error
	RestoreExport(ctx context.Context, sessionID string) error
	GetAll(ctx context.Context) ([]types.ExportInfo, error)
}

type exportService struct {
	sessionStore sessionstore.SessionStore

	fileStore filestore.FileStore

	downloader *seqProxyDownloader

	batchSize    uint64
	workersCount int

	partLength time.Duration
	urlPrefix  string

	tasksChannelSize int

	startCtx context.Context

	authEnabled  bool
	allowedUsers []string
}

const (
	defaultBatchSize        = uint64(10000)
	defaultTasksChannelSize = int(1e6)
	defaultPartLength       = 1 * time.Hour
)

func NewService(
	ctx context.Context,
	cfg config.MassExport,
	sessionStore sessionstore.SessionStore,
	fileStore filestore.FileStore,
	client seqdb.Client,
) (Service, error) {
	batchSize := defaultBatchSize
	if cfg.BatchSize > 0 {
		batchSize = cfg.BatchSize
	}

	if cfg.WorkersCount <= 0 {
		return nil, errors.New("no export workers")
	}

	partLength := defaultPartLength
	if cfg.PartLength > 0 {
		err := checkPartLength(cfg.PartLength)
		if err != nil {
			return nil, fmt.Errorf("check part length: %w", err)
		}

		partLength = cfg.PartLength
	}

	authEnabled := len(cfg.AllowedUsers) > 0
	if !authEnabled {
		logger.Warn("mass exports allowed for all users")
	}

	if cfg.URLPrefix == "" {
		return nil, fmt.Errorf("empty URL prefix")
	}

	tasksChannelSize := defaultTasksChannelSize
	if cfg.TasksChannelSize != 0 {
		if cfg.TasksChannelSize < 0 {
			return nil, fmt.Errorf("negative parts channel capacity: %d", cfg.TasksChannelSize)
		}
		tasksChannelSize = cfg.TasksChannelSize
	}

	downloader := newSeqProxyDownloader(client, *cfg.SeqProxyDownloader)

	return &exportService{
		sessionStore: sessionStore,

		fileStore: fileStore,

		downloader: downloader,

		batchSize:    batchSize,
		workersCount: cfg.WorkersCount,

		partLength: partLength,
		urlPrefix:  cfg.URLPrefix,

		startCtx: ctx,

		tasksChannelSize: tasksChannelSize,

		authEnabled:  authEnabled,
		allowedUsers: cfg.AllowedUsers,
	}, nil
}

func (s *exportService) StartExport(ctx context.Context, req types.StartExportRequest) (types.StartExportResponse, error) {
	userName, err := s.auth(ctx)
	if err != nil {
		return types.StartExportResponse{}, err
	}

	from := floor(req.From, s.partLength)
	to := ceil(req.To, s.partLength)

	if !(req.Window <= s.partLength) {
		return types.StartExportResponse{}, fmt.Errorf("'window' is larger then part length (%s)", s.partLength)
	}

	curTime := time.Now().UnixMilli()
	jobID := fmt.Sprintf("%s_%d", req.Name, curTime)
	sessionID := fmt.Sprintf("job#%s#export#%s", userName, jobID)
	fileStorePathPrefix := fmt.Sprintf("%s/%s", userName, jobID)

	partsCount := int(to.Sub(from) / s.partLength)

	err = s.sessionStore.StartExport(ctx, sessionID, types.ExportInfo{
		ID:     sessionID,
		UserID: userName,

		Status: types.ExportStatusStart,

		CreatedAt: time.Now(),
		StartedAt: time.Now(),

		PartIsUploaded: make([]bool, partsCount),

		FileStorePathPrefix: fileStorePathPrefix,

		From:       from,
		To:         to,
		Query:      req.Query,
		Window:     req.Window,
		BatchSize:  s.batchSize,
		PartLength: s.partLength,
	})

	if err != nil {
		return types.StartExportResponse{}, fmt.Errorf("start export: %w", err)
	}

	go s.exportAllParts(s.startCtx, sessionID)

	return types.StartExportResponse{
		SessionID: sessionID,
	}, nil
}

func (s *exportService) CheckExport(ctx context.Context, sessionID string) (types.ExportInfo, error) {
	if _, err := s.auth(ctx); err != nil {
		return types.ExportInfo{}, err
	}

	info, err := s.sessionStore.CheckExport(ctx, sessionID)
	if err != nil {
		return types.ExportInfo{}, err
	}

	s.setLinks(&info)

	return info, err
}

func (s *exportService) CancelExport(ctx context.Context, sessionID string) error {
	if _, err := s.auth(ctx); err != nil {
		return err
	}

	return s.sessionStore.CancelExport(ctx, sessionID)
}

func (s *exportService) RestoreExport(ctx context.Context, sessionID string) error {
	if _, err := s.auth(ctx); err != nil {
		return err
	}

	info, err := s.sessionStore.CheckExport(ctx, sessionID)
	if err != nil {
		return err
	}

	if info.Status != types.ExportStatusStart {
		return fmt.Errorf(
			"export '%s' must have status '%s' but actual status is '%s'",
			sessionID, types.ExportStatusStart, info.Status,
		)
	}

	curActiveExport, err := s.sessionStore.GetCurActiveExport(ctx)
	if err == nil {
		return fmt.Errorf("has active export: %s", curActiveExport)
	}

	if !errors.Is(err, sessionstore.ErrNoActiveExport) {
		return fmt.Errorf("get current active export: %w", err)
	}

	err = s.sessionStore.Lock(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("lock: %w", err)
	}

	go s.exportAllParts(s.startCtx, sessionID)
	return nil
}

func (s *exportService) GetAll(ctx context.Context) ([]types.ExportInfo, error) {
	if _, err := s.auth(ctx); err != nil {
		return nil, err
	}

	values, err := s.sessionStore.GetAllExports(ctx)
	if err != nil {
		return nil, err
	}

	for i := range values {
		s.setLinks(&values[i])
	}

	return values, nil
}

func (s *exportService) auth(ctx context.Context) (string, error) {
	if !s.authEnabled {
		return "anonymous", nil
	}

	user, err := types.GetUserKey(ctx)
	if err != nil {
		return "", err
	}

	if slices.Index(s.allowedUsers, user) == -1 {
		return "", fmt.Errorf("%w: user '%s' not allowed", types.ErrPermissionDenied, user)
	}

	return user, nil
}

func (s *exportService) getFileLink(fileStorePath string) string {
	return fmt.Sprintf("%s/%s", s.urlPrefix, fileStorePath)
}

func getLinksPath(fileStorePathPrefix string) string {
	return fmt.Sprintf("%s/links", fileStorePathPrefix)
}

func (s *exportService) setLinks(info *types.ExportInfo) {
	linksPath := getLinksPath(info.FileStorePathPrefix)
	info.Links = s.getFileLink(linksPath)
}
