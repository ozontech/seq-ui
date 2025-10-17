package errorgroups

import (
	"context"
	"fmt"
	"math"
	"slices"

	"golang.org/x/sync/errgroup"

	"github.com/ozontech/seq-ui/internal/app/config"
	"github.com/ozontech/seq-ui/internal/app/types"
	repositorych "github.com/ozontech/seq-ui/internal/pkg/repository_ch"
)

type Service struct {
	repo           repositorych.Repository
	logTagsMapping config.LogTagsMapping
}

func New(repo repositorych.Repository, logTagsMapping config.LogTagsMapping) *Service {
	return &Service{
		repo:           repo,
		logTagsMapping: logTagsMapping,
	}
}

func (s *Service) GetErrorGroups(
	ctx context.Context,
	req types.GetErrorGroupsRequest,
) ([]types.ErrorGroup, uint64, error) {
	if req.Service == "" {
		return nil, 0, types.NewErrInvalidRequestField("'service' must not be empty")
	}

	const defaultLimit uint32 = 25
	if req.Limit == 0 {
		req.Limit = defaultLimit
	}

	group, groupCtx := errgroup.WithContext(ctx)

	var groups []types.ErrorGroup
	group.Go(func() error {
		var err error
		groups, err = s.repo.GetErrorGroups(groupCtx, req)
		return err
	})

	var total uint64
	if req.WithTotal {
		group.Go(func() error {
			var err error
			total, err = s.repo.GetErrorGroupsCount(groupCtx, req)
			return err
		})
	}

	err := group.Wait()
	if err != nil {
		return nil, 0, fmt.Errorf("get error groups failed: %w", err)
	}

	return groups, total, err
}

func (s *Service) GetHist(
	ctx context.Context,
	req types.GetErrorHistRequest,
) ([]types.ErrorHistBucket, error) {
	if req.Service == "" {
		return nil, types.NewErrInvalidRequestField("'service' must not be empty")
	}

	return s.repo.GetErrorHist(ctx, req)
}

func (s *Service) GetDetails(
	ctx context.Context,
	req types.GetErrorGroupDetailsRequest,
) (types.ErrorGroupDetails, error) {
	details := types.ErrorGroupDetails{}

	if req.Service == "" {
		return details, types.NewErrInvalidRequestField("'service' must not be empty")
	}
	if req.GroupHash == 0 {
		return details, types.NewErrInvalidRequestField("'group_hash' must not be empty")
	}

	// fast way without calc distributions
	if req.IsFullyFilled() {
		details, err := s.repo.GetErrorDetails(ctx, req)
		if err != nil {
			return details, fmt.Errorf("get error details failed: %w", err)
		}

		if details.SeenTotal > 0 {
			details.Distributions = types.ErrorGroupDistributions{
				ByEnv:     []types.ErrorGroupDistribution{{Value: *req.Env, Percent: 100}},
				ByRelease: []types.ErrorGroupDistribution{{Value: *req.Release, Percent: 100}},
			}
		}
		return details, nil
	}

	group, groupCtx := errgroup.WithContext(ctx)

	group.Go(func() error {
		var err error
		details, err = s.repo.GetErrorDetails(groupCtx, req)
		return err
	})

	var counts types.ErrorGroupCounts
	group.Go(func() error {
		var err error
		counts, err = s.repo.GetErrorCounts(ctx, req)
		return err
	})

	err := group.Wait()
	if err != nil {
		return details, fmt.Errorf("get error details failed: %w", err)
	}

	calcDistribution := func(count types.ErrorGroupCount) []types.ErrorGroupDistribution {
		distr := make([]types.ErrorGroupDistribution, 0, len(count))
		for v, c := range count {
			percent := (float64(c) / float64(details.SeenTotal)) * float64(100)
			distr = append(distr, types.ErrorGroupDistribution{
				Value:   v,
				Percent: uint64(math.Round(percent)),
			})
		}
		slices.SortStableFunc(distr, func(a, b types.ErrorGroupDistribution) int {
			return int(b.Percent) - int(a.Percent)
		})
		return distr
	}

	details.Distributions = types.ErrorGroupDistributions{
		ByEnv:     calcDistribution(counts.ByEnv),
		ByRelease: calcDistribution(counts.ByRelease),
	}

	clearLogTags := func(filter *string, mapping []string) {
		if filter == nil || *filter == "" {
			for _, v := range mapping {
				delete(details.LogTags, v)
			}
		}
	}

	// remove tags if they are not included in request
	clearLogTags(req.Release, s.logTagsMapping.Release)
	clearLogTags(req.Env, s.logTagsMapping.Env)

	return details, nil
}

func (s *Service) GetReleases(
	ctx context.Context,
	req types.GetErrorGroupReleasesRequest,
) ([]string, error) {
	if req.Service == "" {
		return nil, types.NewErrInvalidRequestField("'service' must not be empty")
	}

	return s.repo.GetErrorReleases(ctx, req)
}

func (s *Service) GetServices(
	ctx context.Context,
	req types.GetServicesRequest,
) ([]string, error) {
	return s.repo.GetServices(ctx, req)
}
