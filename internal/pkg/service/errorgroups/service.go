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

const (
	defaultLimit uint32 = 25
)

type Service interface {
	GetErrorGroups(context.Context, types.GetErrorGroupsRequest) ([]types.ErrorGroup, uint64, error)
	GetNewErrorGroups(context.Context, types.GetErrorGroupsRequest) ([]types.ErrorGroup, uint64, error)
	GetTopErrorGroups(context.Context, types.GetTopErrorGroupsRequest) ([]types.TopErrorGroup, uint64, error)

	GetDetails(context.Context, types.GetErrorGroupDetailsRequest) (types.ErrorGroupDetails, error)
	GetHist(context.Context, types.GetErrorHistRequest) ([]types.ErrorHistBucket, error)

	GetServices(context.Context, types.GetServicesRequest) ([]string, error)
	GetReleases(context.Context, types.GetReleasesRequest) ([]string, error)

	DiffByReleases(context.Context, types.DiffByReleasesRequest) ([]types.DiffGroup, uint64, error)
}

type service struct {
	repo           repositorych.Repository
	logTagsMapping config.LogTagsMapping
}

func New(repo repositorych.Repository, logTagsMapping config.LogTagsMapping) Service {
	return &service{
		repo:           repo,
		logTagsMapping: logTagsMapping,
	}
}

func (s *service) GetErrorGroups(
	ctx context.Context,
	req types.GetErrorGroupsRequest,
) ([]types.ErrorGroup, uint64, error) {
	return getErrorGroups(ctx, req, s.repo.GetErrorGroups, s.repo.GetErrorGroupsTotal)
}

func (s *service) GetNewErrorGroups(
	ctx context.Context,
	req types.GetErrorGroupsRequest,
) ([]types.ErrorGroup, uint64, error) {
	// If the release and duration are not specified,
	// then we are looking for errors for all time and releases.
	// In this case, we believe that there are no new errors.
	if (req.Release == nil || *req.Release == "") &&
		(req.Duration == nil || *req.Duration == 0) {
		return nil, 0, nil
	}

	return getErrorGroups(ctx, req, s.repo.GetNewErrorGroups, s.repo.GetNewErrorGroupsTotal)
}

func getErrorGroups(
	ctx context.Context,
	req types.GetErrorGroupsRequest,
	groupsFn func(ctx context.Context, req types.GetErrorGroupsRequest) ([]types.ErrorGroup, error),
	countFn func(ctx context.Context, req types.GetErrorGroupsRequest) (uint64, error),
) ([]types.ErrorGroup, uint64, error) {
	if req.Service == "" {
		return nil, 0, types.NewErrInvalidRequestField("'service' must not be empty")
	}

	if req.Limit == 0 {
		req.Limit = defaultLimit
	}

	eg, ctx := errgroup.WithContext(ctx)

	var groups []types.ErrorGroup
	eg.Go(func() error {
		var err error
		groups, err = groupsFn(ctx, req)
		return err
	})

	var total uint64
	if req.WithTotal {
		eg.Go(func() error {
			var err error
			total, err = countFn(ctx, req)
			return err
		})
	}

	err := eg.Wait()
	if err != nil {
		return nil, 0, fmt.Errorf("get error groups failed: %w", err)
	}

	return groups, total, err
}

func (s *service) GetTopErrorGroups(
	ctx context.Context,
	req types.GetTopErrorGroupsRequest,
) ([]types.TopErrorGroup, uint64, error) {
	if req.Limit == 0 {
		req.Limit = defaultLimit
	}

	eg, ctx := errgroup.WithContext(ctx)

	var groups []types.TopErrorGroup
	eg.Go(func() error {
		var err error
		groups, err = s.repo.GetTopErrorGroups(ctx, req)
		return err
	})

	var total uint64
	if req.WithTotal {
		eg.Go(func() error {
			var err error
			total, err = s.repo.GetTopErrorGroupsTotal(ctx, req)
			return err
		})
	}

	err := eg.Wait()
	if err != nil {
		return nil, 0, fmt.Errorf("get top error groups failed: %w", err)
	}

	return groups, total, err
}

func (s *service) GetHist(
	ctx context.Context,
	req types.GetErrorHistRequest,
) ([]types.ErrorHistBucket, error) {
	return s.repo.GetErrorHist(ctx, req)
}

func (s *service) GetDetails(
	ctx context.Context,
	req types.GetErrorGroupDetailsRequest,
) (types.ErrorGroupDetails, error) {
	details := types.ErrorGroupDetails{}

	if req.GroupHash == 0 {
		return details, types.NewErrInvalidRequestField("'group_hash' must not be empty")
	}

	// fast way without calc distributions
	if req.IsFullyFilled() {
		details, err := s.repo.GetErrorDetails(ctx, req)
		if err != nil {
			return details, fmt.Errorf("get error details failed: %w", err)
		}
		return details, nil
	}

	eg, groupCtx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		var err error
		details, err = s.repo.GetErrorDetails(groupCtx, req)
		return err
	})

	var counts types.ErrorGroupCounts
	eg.Go(func() error {
		var err error
		counts, err = s.repo.GetErrorCounts(ctx, req)
		return err
	})

	err := eg.Wait()
	if err != nil {
		return details, fmt.Errorf("get error details failed: %w", err)
	}

	calcDistribution := func(count types.ErrorGroupCount, filter *string) []types.ErrorGroupDistribution {
		// calculate distribution for unfiltered columns only.
		// if filter is set, its distribution is always 100%.
		if len(count) == 0 || (filter != nil && *filter != "") {
			return nil
		}

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

	if details.SeenTotal > 0 {
		details.Distributions = types.ErrorGroupDistributions{
			ByEnv:     calcDistribution(counts.ByEnv, req.Env),
			BySource:  calcDistribution(counts.BySource, req.Source),
			ByService: calcDistribution(counts.ByService, req.Service),
			ByRelease: calcDistribution(counts.ByRelease, req.Release),
		}
	}

	clearLogTags := func(filter *string, mapping []string) {
		if filter == nil || *filter == "" {
			for _, v := range mapping {
				delete(details.LogTags, v)
			}
		}
	}

	// remove tags if they are not included in request
	clearLogTags(req.Env, s.logTagsMapping.Env)
	clearLogTags(req.Service, s.logTagsMapping.Service)
	clearLogTags(req.Release, s.logTagsMapping.Release)

	return details, nil
}

func (s *service) GetServices(
	ctx context.Context,
	req types.GetServicesRequest,
) ([]string, error) {
	return s.repo.GetServices(ctx, req)
}

func (s *service) GetReleases(
	ctx context.Context,
	req types.GetReleasesRequest,
) ([]string, error) {
	if req.Service == "" {
		return nil, types.NewErrInvalidRequestField("'service' must not be empty")
	}

	return s.repo.GetReleases(ctx, req)
}

func (s *service) DiffByReleases(
	ctx context.Context,
	req types.DiffByReleasesRequest,
) ([]types.DiffGroup, uint64, error) {
	if req.Service == "" {
		return nil, 0, types.NewErrInvalidRequestField("'service' must be non-empty")
	}
	if len(req.Releases) < 2 {
		return nil, 0, types.NewErrInvalidRequestField("length of'releases' must be at least 2")
	}
	if slices.Contains(req.Releases, "") {
		return nil, 0, types.NewErrInvalidRequestField("each element in 'releases' must be non-empty")
	}

	if req.Limit == 0 {
		req.Limit = defaultLimit
	}

	eg, ctx := errgroup.WithContext(ctx)

	var groups []types.DiffGroup
	eg.Go(func() error {
		var err error
		groups, err = s.repo.DiffByReleases(ctx, req)
		return err
	})

	var total uint64
	if req.WithTotal {
		eg.Go(func() error {
			var err error
			total, err = s.repo.DiffByReleasesTotal(ctx, req)
			return err
		})
	}

	err := eg.Wait()
	if err != nil {
		return nil, 0, fmt.Errorf("diff by releases failed: %w", err)
	}

	return groups, total, err
}
