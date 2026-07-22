package v2

import (
	"fmt"
	"time"
)

// Actual configuration scheme
// version:
// server:
//   http:
//     addr:
//     read_timeout:
//     read_header_timeout:
//     write_timeout:
//     cors:
//       allowed_origins:
//       allowed_methods:
//       allowed_headers:
//       exposed_headers:
//       allow_credentials:
//       max_age:
//       options_passthrough:
//   grpc:
//     addr:
//     connection_timeout:
//   debug:
//     addr:
//   auth:
//     oidc:
//       skip_verify:
//       auth_urls:
//       root_ca:
//       ca_cert:
//       private_key:
//       ssl_skip_verify:
//       allowed_clients:
//       cache_secret_key:
//     jwt:
//       secret_key:
//   rate_limiters:
//     <api_name>:
//       default:
//         rate_per_sec:
//         max_burst:
//         store_max_keys:
//         per_handler:
//       spec_users:
//         <username>:
//           rate_per_sec:
//           max_burst:
//           store_max_keys:
//           per_handler:
// clients:
//   seq_db:
//     id:
//     timeout:
//     avg_doc_size:
//     addrs:
//     request_retries:
//     initial_retry_backoff:
//     max_retry_backoff:
//     client_mode:
//     grpc_keepalive_params:
//       time:
//       timeout:
//       permit_without_stream:
//     download_params:
//       delay:
//       initial_retry_backoff:
//       max_retry_backoff:
//   clickhouse:
//     id:
//     addrs:
//     database:
//     username:
//     password:
//     sharded:
//     dial_timeout:
//     read_timeout:
// handlers:
//   seq_api:
//   	options:
//       max_search_limit:
//       max_search_total_limit:
//       max_search_offset_limit:
//       max_export_limit:
//       seq_cli_max_search_limit:
//       max_parallel_export_requests:
//       max_aggregations_per_request:
//       max_buckets_per_aggregation_ts:
//       events_cache_ttl:
//       pinned_fields:
//         name:
//         type:
//       system_fields:
//         name:
//         type:
//       logs_lifespan_cache_key:
//       logs_lifespan_cache_ttl:
//       fields_cache_ttl:
//       masking:
//         masks:
//           re:
//           groups:
//           mode:
//           replace_word:
//           process_fields:
//           ignore_fields:
//           field_filters:
//             condition:
//             filters:
//               field:
//               mode:
//               values:
//         process_fields:
//         ignore_fields:
//     envs:
//       <env_name>:
//         seq_db_id:
//         options:
//           max_search_limit:
//           max_search_total_limit:
//           max_search_offset_limit:
//           max_export_limit:
//           seq_cli_max_search_limit:
//           max_parallel_export_requests:
//           max_aggregations_per_request:
//           max_buckets_per_aggregation_ts:
//           events_cache_ttl:
//           pinned_fields:
//             name:
//             type:
//           system_fields:
//             name:
//             type:
//           logs_lifespan_cache_key:
//           logs_lifespan_cache_ttl:
//           fields_cache_ttl:
//           masking:
//             masks:
//               re:
//               groups:
//               mode:
//               replace_word:
//               process_fields:
//               ignore_fields:
//               field_filters:
//                 condition:
//                 filters:
//                   field:
//                   mode:
//                   values:
//             process_fields:
//             ignore_fields:
//     default_env:
//   error_groups:
//     log_tags_mapping:
//       env:
//       service:
//       release:
//     query_filter:
//       <ch_column>:
//   mass_export:
//     batch_size:
//     workers_count:
//     tasks_channel_size:
//     part_length:
//     url_prefix:
//     allowed_users:
//     file_store:
//       s3:
//         endpoint:
//         access_key_id:
//         secret_access_key:
//         bucket_name:
//         enable_ssl:
//     session_store:
//       redis_id:
//       export_lifetime:
//     download_params:
//       delay:
//       initial_retry_backoff:
//       max_retry_backoff:
//   async_search:
//     admin_users:
//     list_query_length_limit:
// db:
//   name:
//   host:
//   port:
//   pass:
//   user:
//   request_timeout:
//   connection_pool_capacity:
//   use_prepared_statements:
// cache:
//   inmemory:
//     id:
//     num_counters:
//     max_cost:
//     buffer_items:
//   redis:
//     id:
//     with_inmem_id:
//     addr:
//     username:
//     password:
//     timeout:
//     max_retries:
//     min_retry_backoff:
//     max_retry_backoff:

const (
	DefaultSeqDBClientID     = "default"
	DefaultInmemCacheID      = "seqapi"
	DefaultRedisID           = "default"
	DefaultMassExportRedisID = "mass_export"

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

type Config struct {
	Version  int       `yaml:"version"`
	Server   *Server   `yaml:"server"`
	Clients  *Clients  `yaml:"clients"`
	Handlers *Handlers `yaml:"handlers"`
	DB       *DB       `yaml:"db"`
	Cache    *Cache    `yaml:"cache"`
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
	ID          string `yaml:"id"`
	NumCounters int64  `yaml:"num_counters"`
	MaxCost     int64  `yaml:"max_cost"`
	BufferItems int64  `yaml:"buffer_items"`
}

type Redis struct {
	ID              string        `yaml:"id"`
	WithInmemID     string        `yaml:"with_inmem_id"`
	Addr            string        `yaml:"addr"`
	Username        string        `yaml:"username"`
	Password        string        `yaml:"password"`
	Timeout         time.Duration `yaml:"timeout"`
	MaxRetries      int           `yaml:"max_retries"`
	MinRetryBackoff time.Duration `yaml:"min_retry_backoff"`
	MaxRetryBackoff time.Duration `yaml:"max_retry_backoff"`
}

type Cache struct {
	Inmemory []InmemoryCache `yaml:"inmemory"`
	Redis    []Redis         `yaml:"redis"`
}

type S3 struct {
	Endpoint        string `yaml:"endpoint"`
	AccessKeyID     string `yaml:"access_key_id"`
	SecretAccessKey string `yaml:"secret_access_key"`
	BucketName      string `yaml:"bucket_name"`
	EnableSSl       bool   `yaml:"enable_ssl"`
}

type DownloadParams struct {
	Delay               time.Duration `yaml:"delay"`
	InitialRetryBackoff time.Duration `yaml:"initial_retry_backoff"`
	MaxRetryBackoff     time.Duration `yaml:"max_retry_backoff"`
}

type SessionStore struct {
	RedisID        string        `yaml:"redis_id"`
	ExportLifetime time.Duration `yaml:"export_lifetime"`
}

type FileStore struct {
	S3 *S3 `yaml:"s3"`
}

type MassExport struct {
	BatchSize        uint64          `yaml:"batch_size"`
	WorkersCount     int             `yaml:"workers_count"`
	TasksChannelSize int             `yaml:"tasks_channel_size"`
	PartLength       time.Duration   `yaml:"part_length"`
	URLPrefix        string          `yaml:"url_prefix"`
	AllowedUsers     []string        `yaml:"allowed_users"`
	FileStore        *FileStore      `yaml:"file_store"`
	SessionStore     *SessionStore   `yaml:"session_store"`
	DownloadParams   *DownloadParams `yaml:"download_params"`
}

type HTTP struct {
	Addr              string        `yaml:"addr"`
	ReadHeaderTimeout time.Duration `yaml:"read_header_timeout"`
	ReadTimeout       time.Duration `yaml:"read_timeout"`
	WriteTimeout      time.Duration `yaml:"write_timeout"`
	CORS              *CORS         `yaml:"cors"`
}

type GRPC struct {
	Addr              string        `yaml:"addr"`
	ConnectionTimeout time.Duration `yaml:"connection_timeout"`
}

type Debug struct {
	Addr string `yaml:"addr"`
}

type JWT struct {
	SecretKey string `yaml:"secret_key"`
}

type Auth struct {
	OIDC *OIDC `yaml:"oidc"`
	JWT  *JWT  `yaml:"jwt"`
}

type Server struct {
	HTTP         HTTP              `yaml:"http"`
	GRPC         GRPC              `yaml:"grpc"`
	Debug        Debug             `yaml:"debug"`
	Auth         *Auth             `yaml:"auth"`
	RateLimiters ApiToRateLimiters `yaml:"rate_limiters"`
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
	DownloadParams      *DownloadParams      `yaml:"download_params"`
}

type CHClient struct {
	ID          string        `yaml:"id"`
	Addrs       []string      `yaml:"addrs"`
	Database    string        `yaml:"database"`
	Username    string        `yaml:"username"`
	Password    string        `yaml:"password"`
	Sharded     bool          `yaml:"sharded"`
	DialTimeout time.Duration `yaml:"dial_timeout"`
	ReadTimeout time.Duration `yaml:"read_timeout"`
}

type Clients struct {
	SeqDB      []SeqDBClient `yaml:"seq_db"`
	ClickHouse []CHClient    `yaml:"clickhouse"`
}

type Handlers struct {
	SeqAPI      SeqAPI      `yaml:"seq_api"`
	ErrorGroups ErrorGroups `yaml:"error_groups"`
	MassExport  *MassExport `yaml:"mass_export"`
	AsyncSearch AsyncSearch `yaml:"async_search"`
}

type Field struct {
	Name string `yaml:"name"`
	Type string `yaml:"type"`
}

type SeqAPI struct {
	Options    SeqAPIOptions        `yaml:"options"`
	Envs       map[string]SeqAPIEnv `yaml:"envs"`
	DefaultEnv string               `yaml:"default_env"`
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
	PinnedFields               []Field       `yaml:"pinned_fields"`
	SystemFields               []Field       `yaml:"system_fields"`
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
	Service []string `yaml:"service"`
	Env     []string `yaml:"env"`
}

type ErrorGroups struct {
	LogTagsMapping LogTagsMapping    `yaml:"log_tags_mapping"`
	QueryFilter    map[string]string `yaml:"query_filter"`
}

type AsyncSearch struct {
	AdminUsers           []string `yaml:"admin_users"`
	ListQueryLengthLimit int      `yaml:"list_query_length_limit"`
}

func Normalize(cfg *Config) error {
	if len(cfg.Clients.SeqDB) == 0 {
		return fmt.Errorf("clients.seq_db must contain at least one client")
	}

	seqDBIDs := make(map[string]struct{}, len(cfg.Clients.SeqDB))
	for i := range cfg.Clients.SeqDB {
		c := &cfg.Clients.SeqDB[i]
		if c.ID == "" {
			return fmt.Errorf("seq_db client ID cannot be empty")
		}
		if _, ok := seqDBIDs[c.ID]; ok {
			return fmt.Errorf("duplicate seq_db client ID: %s", c.ID)
		}

		seqDBIDs[c.ID] = struct{}{}

		if c.ClientMode == "" {
			c.ClientMode = ProxyClientModeGRPC
		} else if c.ClientMode != ProxyClientModeGRPC {
			return fmt.Errorf("invalid clients.seq_db[%s].client_mode: %q (allowed: %q)", c.ID, c.ClientMode, ProxyClientModeGRPC)
		}

		if c.GRPCKeepaliveParams != nil {
			if c.GRPCKeepaliveParams.Time < minGRPCKeepaliveTime {
				c.GRPCKeepaliveParams.Time = minGRPCKeepaliveTime
			}
			if c.GRPCKeepaliveParams.Timeout < minGRPCKeepaliveTimeout {
				c.GRPCKeepaliveParams.Timeout = minGRPCKeepaliveTimeout
			}
		}
	}

	chIDs := make(map[string]struct{}, len(cfg.Clients.ClickHouse))
	for i := range cfg.Clients.ClickHouse {
		ch := &cfg.Clients.ClickHouse[i]
		if ch.ID == "" {
			return fmt.Errorf("clickhouse client ID cannot be empty")
		}
		if _, ok := chIDs[ch.ID]; ok {
			return fmt.Errorf("duplicate clickhouse client ID: %s", ch.ID)
		}

		chIDs[ch.ID] = struct{}{}

		if ch.DialTimeout <= 0 {
			ch.DialTimeout = defaultClickHouseDialTimeout
		}
		if ch.ReadTimeout <= 0 {
			ch.ReadTimeout = defaultClickHouseReadTimeout
		}
	}

	inmemIDs := make(map[string]struct{}, len(cfg.Cache.Inmemory))
	for i := range cfg.Cache.Inmemory {
		inm := &cfg.Cache.Inmemory[i]
		if inm.ID == "" {
			return fmt.Errorf("inmemory cache ID cannot be empty")
		}
		if _, ok := inmemIDs[inm.ID]; ok {
			return fmt.Errorf("duplicate inmemory cache ID: %s", inm.ID)
		}

		inmemIDs[inm.ID] = struct{}{}

		if inm.NumCounters <= 0 {
			inm.NumCounters = defaultInmemCacheNumCounters
		}
		if inm.MaxCost <= 0 {
			inm.MaxCost = defaultInmemCacheMaxCost
		}
		if inm.BufferItems <= 0 {
			inm.BufferItems = defaultInmemCacheBufferItems
		}
	}

	redisIDs := make(map[string]struct{}, len(cfg.Cache.Redis))
	for i := range cfg.Cache.Redis {
		r := &cfg.Cache.Redis[i]
		if r.ID == "" {
			return fmt.Errorf("redis cache ID cannot be empty")
		}
		if _, ok := redisIDs[r.ID]; ok {
			return fmt.Errorf("duplicate redis cache ID: %s", r.ID)
		}

		redisIDs[r.ID] = struct{}{}

		if r.WithInmemID != "" {
			if _, ok := inmemIDs[r.WithInmemID]; !ok {
				return fmt.Errorf("redis cache %q references unknown inmem cache id %q", r.ID, r.WithInmemID)
			}
		}
	}

	if cfg.DB != nil && cfg.DB.UsePreparedStatements == nil {
		cfg.DB.UsePreparedStatements = new(bool)
		*cfg.DB.UsePreparedStatements = true
	}

	if cfg.Handlers.AsyncSearch.ListQueryLengthLimit <= 0 {
		cfg.Handlers.AsyncSearch.ListQueryLengthLimit = defaultAsyncSearchListQueryLengthLimit
	}

	setSeqAPIOptionsDefaults(&cfg.Handlers.SeqAPI.Options)

	if len(cfg.Handlers.SeqAPI.Envs) > 0 {
		if cfg.Handlers.SeqAPI.DefaultEnv == "" {
			return fmt.Errorf("default_env must be specified when using envs")
		}

		if _, exists := cfg.Handlers.SeqAPI.Envs[cfg.Handlers.SeqAPI.DefaultEnv]; !exists {
			return fmt.Errorf("default_env '%s' not found in seq_api.envs", cfg.Handlers.SeqAPI.DefaultEnv)
		}

		for envName, envConfig := range cfg.Handlers.SeqAPI.Envs {
			if _, ok := seqDBIDs[envConfig.SeqDB]; !ok {
				return fmt.Errorf("client '%s' for env '%s' not found", envConfig.SeqDB, envName)
			}

			if envConfig.Options == nil {
				envConfig.Options = cfg.Handlers.SeqAPI.SeqAPIOptions
			} else {
				setSeqAPIOptionsDefaults(envConfig.Options)
			}

			cfg.Handlers.SeqAPI.Envs[envName] = envConfig
		}
	}

	if cfg.Handlers.MassExport != nil {
		if cfg.Handlers.MassExport.SessionStore.RedisID == "" {
			return fmt.Errorf("handlers.mass_export.session_store.redis_id cannot be empty")
		}

		if _, ok := redisIDs[cfg.Handlers.MassExport.SessionStore.RedisID]; !ok {
			return fmt.Errorf("unknown handlers.mass_export.session_store.redis_id %q", cfg.Handlers.MassExport.SessionStore.RedisID)
		}
	}
	return nil
}

func setSeqAPIOptionsDefaults(options *SeqAPIOptions) {
	if options.MaxAggregationsPerRequest <= 0 {
		options.MaxAggregationsPerRequest = defaultMaxAggregationsPerRequest
	}
	if options.MaxBucketsPerAggregationTs <= 0 {
		options.MaxBucketsPerAggregationTs = defaultMaxBucketsPerAggregationTs
	}
	if options.MaxParallelExportRequests <= 0 {
		options.MaxParallelExportRequests = defaultMaxParallelExportRequests
	}
	if options.MaxSearchTotalLimit <= 0 {
		options.MaxSearchTotalLimit = defaultMaxSearchTotalLimit
	}
	if options.MaxSearchOffsetLimit <= 0 {
		options.MaxSearchOffsetLimit = defaultMaxSearchOffsetLimit
	}
	if options.MaxExportLimit <= 0 {
		options.MaxExportLimit = defaultMaxExportLimit
	}
	if options.EventsCacheTTL <= 0 {
		options.EventsCacheTTL = defaultEventsCacheTTL
	}
	if options.LogsLifespanCacheKey == "" {
		options.LogsLifespanCacheKey = defaultLogsLifespanCacheKey
	}
	if options.LogsLifespanCacheTTL <= 0 {
		options.LogsLifespanCacheTTL = defaultLogsLifespanCacheTTL
	}
}
