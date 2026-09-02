package types

import (
	"iter"
	"maps"
	"slices"
	"time"
)

type ErrorGroupsOrder int

const (
	OrderFrequent ErrorGroupsOrder = iota
	OrderLatest
	OrderOldest
)

type TimeRange struct {
	Duration time.Duration

	From time.Time
	To   time.Time
}

func (tr *TimeRange) IsAbsolute() bool {
	return tr != nil && !tr.From.IsZero() && !tr.To.IsZero()
}

func (tr *TimeRange) AbsoluteDuration() time.Duration {
	if !tr.IsAbsolute() {
		return 0
	}
	return tr.To.Sub(tr.From)
}

func (tr *TimeRange) IsRelative() bool {
	return tr != nil && tr.Duration != 0
}

func (tr *TimeRange) IsEmpty() bool {
	return tr == nil || !tr.IsAbsolute() && !tr.IsRelative()
}

type ErrorGroupsFilter struct {
	Custom map[string]string
}

func (f *ErrorGroupsFilter) CustomOrdered() iter.Seq2[string, string] {
	keys := slices.Collect(maps.Keys(f.Custom))
	slices.Sort(keys)
	return func(yield func(string, string) bool) {
		for _, key := range keys {
			if !yield(key, f.Custom[key]) {
				return
			}
		}
	}
}

type GetErrorGroupsRequest struct {
	Service   string
	Env       *string
	Source    *string
	Release   *string
	TimeRange *TimeRange
	Filter    *ErrorGroupsFilter
	Limit     uint32
	Offset    uint32
	Order     ErrorGroupsOrder
	WithTotal bool
}

type ErrorGroup struct {
	Hash        uint64
	Source      string
	Message     string
	Count       uint64
	FirstSeenAt time.Time
	LastSeenAt  time.Time
}

type GetTopErrorGroupsRequest struct {
	Env       *string
	Source    *string
	TimeRange *TimeRange
	Limit     uint32
	Offset    uint32
	WithTotal bool
}

type TopErrorGroup struct {
	Hash    uint64
	Source  string
	Message string
	Count   uint64
}

type GetErrorHistRequest struct {
	GroupHash *uint64
	Service   *string
	Env       *string
	Source    *string
	Release   *string
	TimeRange *TimeRange
	Filter    *ErrorGroupsFilter
}

type ErrorHistBucket struct {
	Time  time.Time
	Count uint64
}

type ErrorHist struct {
	Buckets  []ErrorHistBucket
	Interval uint64
}

type GetErrorGroupDetailsRequest struct {
	GroupHash uint64
	Env       *string
	Source    *string
	Service   *string
	Release   *string
	Filter    *ErrorGroupsFilter
}

type ErrorGroupDetails struct {
	Hash          uint64
	Message       string
	SeenTotal     uint64
	FirstSeenAt   time.Time
	LastSeenAt    time.Time
	Source        string
	LogTags       map[string]string
	Distributions ErrorGroupDistributions
}

type ErrorGroupDistribution struct {
	Value   string
	Percent uint64
}

type ErrorGroupDistributions struct {
	ByEnv     []ErrorGroupDistribution
	BySource  []ErrorGroupDistribution
	ByService []ErrorGroupDistribution
	ByRelease []ErrorGroupDistribution
	ByFilter  map[string][]ErrorGroupDistribution
}

type ErrorGroupCount map[string]uint64

type ErrorGroupCounts struct {
	ByEnv     ErrorGroupCount
	BySource  ErrorGroupCount
	ByService ErrorGroupCount
	ByRelease ErrorGroupCount
	ByFilter  map[string]ErrorGroupCount
}

type GetServicesRequest struct {
	Query  string
	Env    *string
	Limit  uint32
	Offset uint32
}

type GetReleasesRequest struct {
	Service string
	Env     *string
}

type DiffByReleasesRequest struct {
	Service   string
	Releases  []string
	Env       *string
	Source    *string
	Limit     uint32
	Offset    uint32
	Order     ErrorGroupsOrder
	WithTotal bool
}

type DiffReleaseInfo struct {
	SeenTotal uint64
}

type DiffGroup struct {
	Hash         uint64
	Message      string
	FirstSeenAt  time.Time
	LastSeenAt   time.Time
	Source       string
	ReleaseInfos map[string]DiffReleaseInfo
}

type GetServiceFiltersRequest struct {
	Service string
	Env     *string
}

type ServiceFilter struct {
	Key    string
	Values []string
}
