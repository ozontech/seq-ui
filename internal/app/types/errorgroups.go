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
	SeenTotal   uint64
	FirstSeenAt time.Time
	LastSeenAt  time.Time
	Source      string
}

type GetErrorHistRequest struct {
	Service   string
	GroupHash *uint64
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
	Service   string
	GroupHash uint64
	Env       *string
	Source    *string
	Release   *string
}

func (r GetErrorGroupDetailsRequest) IsFullyFilled() bool {
	return r.Env != nil && *r.Env != "" &&
		r.Release != nil && *r.Release != ""
}

type ErrorGroupDetails struct {
	GroupHash     uint64
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
	ByRelease []ErrorGroupDistribution
}

type ErrorGroupCount map[string]uint64

type ErrorGroupCounts struct {
	ByEnv     ErrorGroupCount
	ByRelease ErrorGroupCount
}

type GetErrorGroupReleasesRequest struct {
	Service   string
	GroupHash *uint64
	Env       *string
}

type GetServicesRequest struct {
	Query  string
	Env    *string
	Limit  uint32
	Offset uint32
}
