package test

import (
	"fmt"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/ozontech/seq-ui/internal/app/config"
	mock_cache "github.com/ozontech/seq-ui/internal/pkg/cache/mock"
	mock_seqdb "github.com/ozontech/seq-ui/internal/pkg/client/seqdb/mock"
	mock_repo "github.com/ozontech/seq-ui/internal/pkg/repository/mock"
	"github.com/ozontech/seq-ui/pkg/seqapi/v1"
)

type CacheMockArgs struct {
	Key   string
	Value string
	Err   error
}

type Mocks struct {
	SeqDB             *mock_seqdb.MockClient
	Cache             *mock_cache.MockCache
	AsyncSearchesRepo *mock_repo.MockAsyncSearches
	ProfilesRepo      *mock_repo.MockUserProfiles
}

type APITestData struct {
	Cfg   config.SeqAPI
	Mocks Mocks
}

func MakeEvent(id string, countData int, t time.Time) *seqapi.Event {
	e := &seqapi.Event{
		Id:   id,
		Data: make(map[string]string),
		Time: timestamppb.New(t),
	}
	for i := range countData {
		e.Data[fmt.Sprintf("field%d", i+1)] = fmt.Sprintf("val%d", i+1)
	}
	return e
}

func MakeEvents(count int, t time.Time) []*seqapi.Event {
	events := make([]*seqapi.Event, 0, count)
	for i := range count {
		events = append(events, MakeEvent(fmt.Sprintf("test%d", i+1), i+1, t))
	}
	return events
}

func MakeHistogram(bucketCount int) *seqapi.Histogram {
	hist := &seqapi.Histogram{
		Buckets: make([]*seqapi.Histogram_Bucket, 0, bucketCount),
	}
	for i := range bucketCount {
		hist.Buckets = append(hist.Buckets, &seqapi.Histogram_Bucket{
			Key:      uint64(i * 100),
			DocCount: uint64(i + 1),
		})
	}
	return hist
}

type MakeAggOpts struct {
	NotExists int64
	Quantiles []float64
	Ts        []*timestamppb.Timestamp
}

func MakeAggregation(bucketCount int, opts *MakeAggOpts) *seqapi.Aggregation {
	agg := &seqapi.Aggregation{
		Buckets: make([]*seqapi.Aggregation_Bucket, 0, bucketCount),
	}
	for i := range bucketCount {
		v := new(float64)
		*v = float64(i + 1)
		b := &seqapi.Aggregation_Bucket{
			Key:   fmt.Sprintf("test%d", i+1),
			Value: v,
		}
		if opts != nil {
			b.NotExists = opts.NotExists
			if len(opts.Quantiles) > 0 {
				b.Quantiles = opts.Quantiles
			}
			if len(opts.Ts) > 0 {
				b.Ts = opts.Ts[i]
			}
		}
		agg.Buckets = append(agg.Buckets, b)
	}
	return agg
}

func MakeAggregations(aggCount, bucketCount int, opts *MakeAggOpts) []*seqapi.Aggregation {
	aggs := make([]*seqapi.Aggregation, 0, aggCount)
	for range aggCount {
		aggs = append(aggs, MakeAggregation(bucketCount, opts))
	}
	return aggs
}

func SetCfgDefaults(cfg config.SeqAPI) config.SeqAPI {
	for envName, envConfig := range cfg.Envs {
		opts := *envConfig.Options
		if opts.MaxAggregationsPerRequest <= 0 {
			opts.MaxAggregationsPerRequest = 1
		}
		if opts.MaxParallelExportRequests <= 0 {
			opts.MaxParallelExportRequests = 1
		}
		if opts.MaxSearchTotalLimit <= 0 {
			opts.MaxSearchTotalLimit = 1000000
		}
		if opts.MaxSearchOffsetLimit <= 0 {
			opts.MaxSearchOffsetLimit = 1000000
		}

		envConfig.Options = &opts
		cfg.Envs[envName] = envConfig
	}
	return cfg
}
