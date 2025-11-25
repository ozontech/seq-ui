package asyncsearches

import (
	"context"
	"fmt"
	"time"

	"github.com/cenkalti/backoff/v4"
	"go.uber.org/zap"

	"github.com/ozontech/seq-ui/internal/app/types"
	"github.com/ozontech/seq-ui/internal/pkg/client/seqdb"
	"github.com/ozontech/seq-ui/internal/pkg/repository"
	"github.com/ozontech/seq-ui/logger"
	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
)

const (
	backoffInitialInterval = 250 * time.Millisecond
	maxBackoffElapsedTime  = 2 * time.Second

	deleteExpiredAsyncSearchesInterval = 1 * time.Minute
)

type Service struct {
	repo  repository.AsyncSearches
	seqDB seqdb.Client
}

func New(ctx context.Context, repo repository.AsyncSearches, seqDB seqdb.Client) *Service {
	s := &Service{
		repo:  repo,
		seqDB: seqDB,
	}

	go s.deleteExpiredAsyncSearches(ctx)

	return s
}

func (s *Service) StartAsyncSearch(
	ctx context.Context,
	ownerID int64,
	req *seqapi.StartAsyncSearchRequest,
) (*seqapi.StartAsyncSearchResponse, error) {
	resp, err := s.seqDB.StartAsyncSearch(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to start async search: %w", err)
	}

	// we can retry multiple times after the async search saved in seq-db
	err = backoff.Retry(func() error {
		return s.repo.SaveAsyncSearch(ctx, types.SaveAsyncSearchRequest{
			SearchID:  resp.SearchId,
			OwnerID:   ownerID,
			ExpiresAt: time.Now().Add(req.Retention.AsDuration()),
			Meta:      req.Meta,
		})
	}, backoff.NewExponentialBackOff(
		backoff.WithInitialInterval(backoffInitialInterval),
		backoff.WithMaxElapsedTime(maxBackoffElapsedTime),
	))
	if err != nil {
		return nil, fmt.Errorf("failed to save async search: %w", err)
	}

	return resp, nil
}

func (s *Service) DeleteAsyncSearch(
	ctx context.Context,
	ownerID int64,
	req *seqapi.DeleteAsyncSearchRequest,
) (*seqapi.DeleteAsyncSearchResponse, error) {
	searchInfo, err := s.repo.GetAsyncSearchById(ctx, req.SearchId)
	if err != nil {
		return nil, fmt.Errorf("failed to get async search by id: %w", err)
	}

	if searchInfo.OwnerID != ownerID {
		return nil, types.NewErrPermissionDenied("delete async search")
	}

	resp, err := s.seqDB.DeleteAsyncSearch(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to delete async search from seq-db: %w", err)
	}

	// we can retry multiple times after the async search is deleted from seq-db
	err = backoff.Retry(func() error {
		return s.repo.DeleteAsyncSearch(ctx, req.SearchId)
	}, backoff.NewExponentialBackOff(
		backoff.WithInitialInterval(backoffInitialInterval),
		backoff.WithMaxElapsedTime(maxBackoffElapsedTime),
	))
	if err != nil {
		return nil, fmt.Errorf("failed to delete async search from db: %w", err)
	}

	return resp, nil
}

func (s *Service) CancelAsyncSearch(
	ctx context.Context,
	ownerID int64,
	req *seqapi.CancelAsyncSearchRequest,
) (*seqapi.CancelAsyncSearchResponse, error) {
	searchInfo, err := s.repo.GetAsyncSearchById(ctx, req.SearchId)
	if err != nil {
		return nil, fmt.Errorf("failed to get async search by id: %w", err)
	}

	if searchInfo.OwnerID != ownerID {
		return nil, types.NewErrPermissionDenied("cancel async search")
	}

	resp, err := s.seqDB.CancelAsyncSearch(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to cancel async search in seq-db: %w", err)
	}

	return resp, nil
}

func (s *Service) FetchAsyncSearchResult(
	ctx context.Context,
	req *seqapi.FetchAsyncSearchResultRequest,
) (*seqapi.FetchAsyncSearchResultResponse, error) {
	searchInfo, err := s.repo.GetAsyncSearchById(ctx, req.SearchId)
	if err != nil {
		return nil, fmt.Errorf("failed to get async search by id: %w", err)
	}

	resp, err := s.seqDB.FetchAsyncSearchResult(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch search result: %w", err)
	}

	resp.Meta = searchInfo.Meta

	return resp, nil
}

func (s *Service) GetAsyncSearchesList(
	ctx context.Context,
	req *seqapi.GetAsyncSearchesListRequest,
) (*seqapi.GetAsyncSearchesListResponse, error) {
	searches, err := s.repo.GetAsyncSearchesList(ctx, types.GetAsyncSearchesListRequest{
		Owner: req.OwnerName,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get async searches list from db: %w", err)
	}

	ownerNameByID := make(map[string]string, len(searches))
	searchIDs := make([]string, 0, len(searches))
	for _, search := range searches {
		ownerNameByID[search.SearchID] = search.OwnerName
		searchIDs = append(searchIDs, search.SearchID)
	}

	if len(searchIDs) == 0 {
		return &seqapi.GetAsyncSearchesListResponse{}, nil
	}

	resp, err := s.seqDB.GetAsyncSearchesList(ctx, req, searchIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get async searches list from seq-db: %w", err)
	}

	for _, as := range resp.Searches {
		as.OwnerName = ownerNameByID[as.SearchId]
	}

	return resp, nil
}

func (s *Service) deleteExpiredAsyncSearches(ctx context.Context) {
	ticker := time.NewTicker(deleteExpiredAsyncSearchesInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			err := s.repo.DeleteExpiredAsyncSearches(ctx)
			if err != nil {
				logger.Error("DeleteExpiredAsyncSearches error", zap.Error(err))
			}
		}
	}
}
