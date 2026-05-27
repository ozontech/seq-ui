package config

import (
	"fmt"
	"os"
	"time"
)

const (
	DefaultSeqDBClientID = "default"

	ProxyClientModeGRPC = "grpc"

	MaskModeMask    = "mask"
	MaskModeReplace = "replace"
	MaskModeCut     = "cut"

	FieldFilterConditionAnd = "and"
	FieldFilterConditionOr  = "or"
	FieldFilterConditionNot = "not"

	FieldFilterModeEqual    = "equal"
	FieldFilterModeContains = "contains"
	FieldFilterModePrefix   = "prefix"
	FieldFilterModeSuffix   = "suffix"

	minGRPCKeepaliveTime    = 10 * time.Second
	minGRPCKeepaliveTimeout = 1 * time.Second

	defaultAsyncSearchListQueryLengthLimit = 1000

	defaultMaxSearchTotalLimit        = 1000000
	defaultMaxSearchOffsetLimit       = 1000000
	defaultMaxExportLimit             = 100000
	defaultMaxAggregationsPerRequest  = 1
	defaultMaxBucketsPerAggregationTs = 200
	defaultMaxParallelExportRequests  = 1

	defaultInmemCacheNumCounters = 10000000
	defaultInmemCacheMaxCost     = 1000000
	defaultInmemCacheBufferItems = 64

	defaultEventsCacheTTL = 24 * time.Hour

	defaultLogsLifespanCacheKey = "logs_lifespan"
	defaultLogsLifespanCacheTTL = 10 * time.Minute

	defaultClickHouseDialTimeout = 5 * time.Second
	defaultClickHouseReadTimeout = 30 * time.Second
)

func FromFile(cfgPath string) (Config, error) {
	cfgBytes, err := os.ReadFile(cfgPath)
	if err != nil {
		return Config{}, fmt.Errorf("error reading file: %s", err)
	}

	return fromBytes(cfgBytes)
}

func fromBytes(cfgBytes []byte) (Config, error) {

}
