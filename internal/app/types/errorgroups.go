package types

import (
	"time"
)

type ErrorGroupsOrder int

const (
	OrderFrequent ErrorGroupsOrder = iota
	OrderLatest
	OrderOldest
)

type GetErrorGroupsRequest struct {
	Service   string
	Env       *string
	Source    *string
	Release   *string
	Duration  *time.Duration
	Limit     uint32
	Offset    uint32
	Order     ErrorGroupsOrder
	WithTotal bool
}

type ErrorGroup struct {
	Hash        uint64
	Message     string
	Count       uint64
	FirstSeenAt time.Time
	LastSeenAt  time.Time
	Source      string
}

type GetTopErrorGroupsRequest struct {
	Env       *string
	Source    *string
	Duration  *time.Duration
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
	Duration  *time.Duration
}

type ErrorHistBucket struct {
	Time  time.Time
	Count uint64
}

type GetErrorGroupDetailsRequest struct {
	GroupHash uint64
	Env       *string
	Source    *string
	Service   *string
	Release   *string
}

func (r GetErrorGroupDetailsRequest) IsFullyFilled() bool {
	return r.Env != nil && *r.Env != "" &&
		r.Release != nil && *r.Release != "" &&
		r.Service != nil && *r.Service != "" &&
		r.Source != nil && *r.Source != ""
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
}

type ErrorGroupCount map[string]uint64

type ErrorGroupCounts struct {
	ByEnv     ErrorGroupCount
	BySource  ErrorGroupCount
	ByService ErrorGroupCount
	ByRelease ErrorGroupCount
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
