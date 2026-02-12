package config

import (
	"bytes"
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

const (
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

	defaultMaxSearchTotalLimit        = 1000000
	defaultMaxSearchOffsetLimit       = 1000000
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

type Config struct {
	Server   *Server   `yaml:"server"`
	Clients  *Clients  `yaml:"clients"`
	Handlers *Handlers `yaml:"handlers"`
}

type CORS struct {
	AllowedOrigins     []string `yaml:"allowed_origins"`
	AllowedMethods     []string `yaml:"allowed_methods"`
	AllowedHeaders     []string `yaml:"allowed_headers"`
	ExposedHeaders     []string `yaml:"exposed_headers"`
	AllowCredentials   bool     `yaml:"allow_credentials"`
	MaxAge             int      `yaml:"max_age"`
	OptionsPassthrough bool     `yaml:"options_passthrough"`
}

type OIDC struct {
	SkipVerify     bool     `yaml:"skip_verify"`
	AuthURLs       []string `yaml:"auth_urls"`
	RootCA         string   `yaml:"root_ca"`
	CACert         string   `yaml:"ca_cert"`
	PrivateKey     string   `yaml:"private_key"`
	SSLSkipVerify  bool     `yaml:"ssl_skip_verify"`
	AllowedClients []string `yaml:"allowed_clients"`
	CacheSecretKey string   `yaml:"cache_secret_key"`
}

type DB struct {
	Name                   string        `yaml:"name"`
	Host                   string        `yaml:"host"`
	Port                   int64         `yaml:"port"`
	Pass                   string        `yaml:"pass"`
	User                   string        `yaml:"user"`
	RequestTimeout         time.Duration `yaml:"request_timeout"`
	ConnectionPoolCapacity int64         `yaml:"connection_pool_capacity"`
	UsePreparedStatements  *bool         `yaml:"use_prepared_statements,omitempty"`
}

func (db *DB) ConnString() string {
	return fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s pool_max_conns=%d", db.Host, db.Port, db.Name, db.User, db.Pass, db.ConnectionPoolCapacity)
}

type CH struct {
	Addrs       []string      `yaml:"addrs"`
	Database    string        `yaml:"database"`
	Username    string        `yaml:"username"`
	Password    string        `yaml:"password"`
	Sharded     bool          `yaml:"sharded"`
	DialTimeout time.Duration `yaml:"dial_timeout"`
	ReadTimeout time.Duration `yaml:"read_timeout"`
}

type (
	RateLimiter struct {
		RatePerSec   int  `yaml:"rate_per_sec"`
		MaxBurst     int  `yaml:"max_burst"`
		StoreMaxKeys int  `yaml:"store_max_keys"`
		PerHandler   bool `yaml:"per_handler"`
	}

	UserToRateLimiter map[string]RateLimiter

	ApiRateLimiters struct {
		Default      RateLimiter       `yaml:"default"`
		SpecialUsers UserToRateLimiter `yaml:"spec_users"`
	}

	ApiToRateLimiters map[string]ApiRateLimiters
)

type InmemoryCache struct {
	NumCounters int64 `yaml:"num_counters"`
	MaxCost     int64 `yaml:"max_cost"`
	BufferItems int64 `yaml:"buffer_items"`
}

type Redis struct {
	Addr            string        `yaml:"addr"`
	Username        string        `yaml:"username"`
	Password        string        `yaml:"password"`
	Timeout         time.Duration `yaml:"timeout"`
	MaxRetries      int           `yaml:"max_retries"`
	MinRetryBackoff time.Duration `yaml:"min_retry_backoff"`
	MaxRetryBackoff time.Duration `yaml:"max_retry_backoff"`
}

type Cache struct {
	Inmemory InmemoryCache `yaml:"inmemory"`
	Redis    *Redis        `yaml:"redis"`
}

type S3 struct {
	Endpoint        string `yaml:"endpoint"`
	AccessKeyID     string `yaml:"access_key_id"`
	SecretAccessKey string `yaml:"secret_access_key"`
	BucketName      string `yaml:"bucket_name"`
	EnableSSl       bool   `yaml:"enable_ssl"`
}

type SeqProxyDownloader struct {
	Delay               time.Duration `yaml:"delay"`
	InitialRetryBackoff time.Duration `yaml:"initial_retry_backoff"`
	MaxRetryBackoff     time.Duration `yaml:"max_retry_backoff"`
}

type SessionStore struct {
	Redis          Redis         `yaml:"redis"`
	ExportLifetime time.Duration `yaml:"export_lifetime"`
}

type FileStore struct {
	S3 *S3 `yaml:"s3"`
}

type MassExport struct {
	BatchSize          uint64              `yaml:"batch_size"`
	WorkersCount       int                 `yaml:"workers_count"`
	TasksChannelSize   int                 `yaml:"tasks_channel_size"`
	PartLength         time.Duration       `yaml:"part_length"`
	URLPrefix          string              `yaml:"url_prefix"`
	AllowedUsers       []string            `yaml:"allowed_users"`
	FileStore          *FileStore          `yaml:"file_store"`
	SessionStore       *SessionStore       `yaml:"session_store"`
	SeqProxyDownloader *SeqProxyDownloader `yaml:"seq_proxy_downloader"`
}

type Server struct {
	DebugAddr             string            `yaml:"debug_addr"`
	HTTPAddr              string            `yaml:"http_addr"`
	GRPCAddr              string            `yaml:"grpc_addr"`
	CORS                  *CORS             `yaml:"cors"`
	OIDC                  *OIDC             `yaml:"oidc"`
	GRPCConnectionTimeout time.Duration     `yaml:"grpc_connection_timeout"`
	HTTPReadHeaderTimeout time.Duration     `yaml:"http_read_header_timeout"`
	HTTPReadTimeout       time.Duration     `yaml:"http_read_timeout"`
	HTTPWriteTimeout      time.Duration     `yaml:"http_write_timeout"`
	DB                    *DB               `yaml:"db"`
	CH                    *CH               `yaml:"clickhouse"`
	RateLimiters          ApiToRateLimiters `yaml:"rate_limiters"`
	Cache                 Cache             `yaml:"cache"`
	JWTSecretKey          string            `yaml:"jwt_secret_key"`
}

type GRPCKeepaliveParams struct {
	// After a duration of this time if the client doesn't see any activity it
	// pings the server to see if the transport is still alive.
	// If set below 10s, a minimum value of 10s will be used instead.
	Time time.Duration `yaml:"time"`
	// After having pinged for keepalive check, the client waits for a duration
	// of Timeout and if no activity is seen even after that the connection is
	// closed. If set below 1s, a minimum value of 1s will be used instead.
	Timeout time.Duration `yaml:"timeout"`
	// If true, client sends keepalive pings even with no active RPCs. If false,
	// when there are no active RPCs, Time and Timeout will be ignored and no
	// keepalive pings will be sent. False by default.
	PermitWithoutStream bool `yaml:"permit_without_stream"`
}

type SeqDBClient struct {
	ID                  string               `yaml:"id"`
	Timeout             time.Duration        `yaml:"timeout"`
	AvgDocSize          int                  `yaml:"avg_doc_size"`
	Addrs               []string             `yaml:"addrs"`
	RequestRetries      int                  `yaml:"request_retries"`
	InitialRetryBackoff time.Duration        `yaml:"initial_retry_backoff"`
	MaxRetryBackoff     time.Duration        `yaml:"max_retry_backoff"`
	ClientMode          string               `yaml:"client_mode"`
	GRPCKeepaliveParams *GRPCKeepaliveParams `yaml:"grpc_keepalive_params"`
}

type Clients struct {
	SeqDBTimeout        time.Duration        `yaml:"seq_db_timeout"`
	SeqDBAvgDocSize     int                  `yaml:"seq_db_avg_doc_size"`
	SeqDBAddrs          []string             `yaml:"seq_db_addrs"`
	RequestRetries      int                  `yaml:"request_retries"`
	InitialRetryBackoff time.Duration        `yaml:"initial_retry_backoff"`
	MaxRetryBackoff     time.Duration        `yaml:"max_retry_backoff"`
	ProxyClientMode     string               `yaml:"proxy_client_mode"`
	GRPCKeepaliveParams *GRPCKeepaliveParams `yaml:"grpc_keepalive_params"`
	SeqDB               []SeqDBClient        `yaml:"seq_db"`
}

type Handlers struct {
	SeqAPI      SeqAPI      `yaml:"seq_api"`
	ErrorGroups ErrorGroups `yaml:"error_groups"`
	MassExport  *MassExport `yaml:"mass_export"`
	AsyncSearch AsyncSearch `yaml:"async_search"`
}

type PinnedField struct {
	Name string `yaml:"name"`
	Type string `yaml:"type"`
}

type SeqAPI struct {
	SeqAPIOptions `yaml:",inline"`
	Envs          map[string]SeqAPIEnv `yaml:"envs"`
}

type SeqAPIEnv struct {
	SeqDB   string         `yaml:"seq_db_id"`
	Options *SeqAPIOptions `yaml:"options"`
}

type SeqAPIOptions struct {
	MaxSearchLimit             int32         `yaml:"max_search_limit"`
	MaxSearchTotalLimit        int64         `yaml:"max_search_total_limit"`
	MaxSearchOffsetLimit       int32         `yaml:"max_search_offset_limit"`
	MaxExportLimit             int32         `yaml:"max_export_limit"`
	SeqCLIMaxSearchLimit       int           `yaml:"seq_cli_max_search_limit"`
	MaxParallelExportRequests  int           `yaml:"max_parallel_export_requests"`
	MaxAggregationsPerRequest  int           `yaml:"max_aggregations_per_request"`
	MaxBucketsPerAggregationTs int           `yaml:"max_buckets_per_aggregation_ts"`
	EventsCacheTTL             time.Duration `yaml:"events_cache_ttl"`
	PinnedFields               []PinnedField `yaml:"pinned_fields"`
	LogsLifespanCacheKey       string        `yaml:"logs_lifespan_cache_key"`
	LogsLifespanCacheTTL       time.Duration `yaml:"logs_lifespan_cache_ttl"`
	FieldsCacheTTL             time.Duration `yaml:"fields_cache_ttl"`
	Masking                    *Masking      `yaml:"masking"`
}

type Masking struct {
	Masks         []Mask   `yaml:"masks"`
	ProcessFields []string `yaml:"process_fields"`
	IgnoreFields  []string `yaml:"ignore_fields"`
}

type Mask struct {
	Re          string `yaml:"re"`
	Groups      []int  `yaml:"groups"`
	Mode        string `yaml:"mode"`         // "mask" or "replace" or "cut"
	ReplaceWord string `yaml:"replace_word"` // for mode:replace

	ProcessFields []string `yaml:"process_fields"`
	IgnoreFields  []string `yaml:"ignore_fields"`

	FieldFilters *FieldFilterSet `yaml:"field_filters"`
}

type FieldFilter struct {
	Field  string   `yaml:"field"`
	Mode   string   `yaml:"mode"` // "equal" or "contains" or "prefix" or "suffix"
	Values []string `yaml:"values"`
}

type FieldFilterSet struct {
	Condition string        `yaml:"condition"` // "and" or "or" or "not"
	Filters   []FieldFilter `yaml:"filters"`   // max 1 if condition:not
}

type LogTagsMapping struct {
	Release []string `yaml:"release"`
	Env     []string `yaml:"env"`
}

type ErrorGroups struct {
	LogTagsMapping LogTagsMapping    `yaml:"log_tags_mapping"`
	QueryFilter    map[string]string `yaml:"query_filter"`
}

type AsyncSearch struct {
	AdminUsers []string `yaml:"admin_users"`
}

// FromFile parse config from config path.
func FromFile(cfgPath string) (Config, error) {
	cfgBytes, err := os.ReadFile(cfgPath) //nolint:gosec
	if err != nil {
		return Config{}, fmt.Errorf("error reading file: %s", err)
	}

	cfg, err := parse(cfgBytes)
	if err != nil {
		return Config{}, fmt.Errorf("error parsing file: %s", err)
	}

	proxyClientMode := cfg.Clients.ProxyClientMode
	if proxyClientMode == "" {
		cfg.Clients.ProxyClientMode = ProxyClientModeGRPC
	} else if proxyClientMode != ProxyClientModeGRPC {
		return Config{}, fmt.Errorf(
			"invalid value for clients.proxy_client_mode: %q. Allowed values are empty string (defaults to %q) or %q",
			proxyClientMode, ProxyClientModeGRPC, ProxyClientModeGRPC,
		)
	}

	if cfg.Handlers.SeqAPI.MaxAggregationsPerRequest <= 0 {
		cfg.Handlers.SeqAPI.MaxAggregationsPerRequest = defaultMaxAggregationsPerRequest
	}
	if cfg.Handlers.SeqAPI.MaxBucketsPerAggregationTs <= 0 {
		cfg.Handlers.SeqAPI.MaxBucketsPerAggregationTs = defaultMaxBucketsPerAggregationTs
	}
	if cfg.Handlers.SeqAPI.MaxParallelExportRequests <= 0 {
		cfg.Handlers.SeqAPI.MaxParallelExportRequests = defaultMaxParallelExportRequests
	}
	if cfg.Handlers.SeqAPI.MaxSearchTotalLimit <= 0 {
		cfg.Handlers.SeqAPI.MaxSearchTotalLimit = defaultMaxSearchTotalLimit
	}
	if cfg.Handlers.SeqAPI.MaxSearchOffsetLimit <= 0 {
		cfg.Handlers.SeqAPI.MaxSearchOffsetLimit = defaultMaxSearchOffsetLimit
	}
	if cfg.Handlers.SeqAPI.EventsCacheTTL <= 0 {
		cfg.Handlers.SeqAPI.EventsCacheTTL = defaultEventsCacheTTL
	}
	if cfg.Handlers.SeqAPI.LogsLifespanCacheKey == "" {
		cfg.Handlers.SeqAPI.LogsLifespanCacheKey = defaultLogsLifespanCacheKey
	}
	if cfg.Handlers.SeqAPI.LogsLifespanCacheTTL <= 0 {
		cfg.Handlers.SeqAPI.LogsLifespanCacheTTL = defaultLogsLifespanCacheTTL
	}

	if cfg.Server.DB != nil && cfg.Server.DB.UsePreparedStatements == nil {
		cfg.Server.DB.UsePreparedStatements = new(bool)
		*cfg.Server.DB.UsePreparedStatements = true
	}

	if cfg.Server.CH != nil && cfg.Server.CH.DialTimeout <= 0 {
		cfg.Server.CH.DialTimeout = defaultClickHouseDialTimeout
	}

	if cfg.Server.CH != nil && cfg.Server.CH.ReadTimeout <= 0 {
		cfg.Server.CH.ReadTimeout = defaultClickHouseReadTimeout
	}

	if cfg.Clients.GRPCKeepaliveParams != nil {
		if cfg.Clients.GRPCKeepaliveParams.Time < minGRPCKeepaliveTime {
			cfg.Clients.GRPCKeepaliveParams.Time = minGRPCKeepaliveTime
		}
		if cfg.Clients.GRPCKeepaliveParams.Timeout < minGRPCKeepaliveTimeout {
			cfg.Clients.GRPCKeepaliveParams.Timeout = minGRPCKeepaliveTimeout
		}
	}

	if cfg.Server.Cache.Inmemory.NumCounters <= 0 {
		cfg.Server.Cache.Inmemory.NumCounters = defaultInmemCacheNumCounters
	}
	if cfg.Server.Cache.Inmemory.MaxCost <= 0 {
		cfg.Server.Cache.Inmemory.MaxCost = defaultInmemCacheMaxCost
	}
	if cfg.Server.Cache.Inmemory.BufferItems <= 0 {
		cfg.Server.Cache.Inmemory.BufferItems = defaultInmemCacheBufferItems
	}

	return cfg, nil
}

func parse(cfg []byte) (Config, error) {
	result := Config{}

	decoder := yaml.NewDecoder(bytes.NewReader(cfg))
	decoder.KnownFields(true)
	if err := decoder.Decode(&result); err != nil {
		return result, fmt.Errorf("error parsing config: %w", err)
	}

	return result, nil
}
