package config

import (
	"fmt"
	"time"
)

// Configuration scheme
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
//       cache_id:
//       cache_secret_key:
//       skip_verify:
//       auth_urls:
//       tls:
//         root_ca:
//         ca_cert:
//         private_key:
//         insecure:
//       allowed_clients:
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
//     - id:
//       timeout:
//       avg_doc_size:
//       addrs:
//       request_retries:
//       initial_retry_backoff:
//       max_retry_backoff:
//       client_mode:
//       grpc_keepalive_params:
//         time:
//         timeout:
//         permit_without_stream:
//       download_params:
//         delay:
//         initial_retry_backoff:
//         max_retry_backoff:
//   clickhouse:
//     - id:
//       addrs:
//       database:
//       username:
//       password:
//       sharded:
//       dial_timeout:
//       read_timeout:
// handlers:
//   seq_api:
//     seq_db_id:
//     cache_id:
//     redis_id:
//     global_options:
//       events_cache_ttl:
//       pinned_fields:
//         - name:
//           type:
//       system_fields:
//         - name:
//           type:
//       logs_lifespan_cache_ttl:
//       fields_cache_ttl:
//       masking:
//         masks:
//           - re:
//             groups:
//             mode:
//             replace_word:
//             process_fields:
//             ignore_fields:
//             field_filters:
//               condition:
//               filters:
//                 - field:
//                   mode:
//                   values:
//         process_fields:
//         ignore_fields:
//     options:
//       max_search_limit:
//       max_search_total_limit:
//       max_search_offset_limit:
//       max_export_limit:
//       seq_cli_max_search_limit:
//       max_parallel_export_requests:
//       max_aggregations_per_request:
//       max_buckets_per_aggregation_ts:
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
//     default_env:
//   error_groups:
//     ch_id:
//     envs:
//       <env_name>:
//         ch_id:
//     default_env:
//     log_tags_mapping:
//       env:
//       service:
//       release:
//     query_filter:
//       <ch_column>:
//   mass_export:
//     seq_db_id:
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
//     seq_db_id:
//     envs:
//       <env_name>:
//         seq_db_id:
//         list_query_length_limit:
//     default_env:
//     admin_users:
//     list_query_length_limit:
//   admin:
//     redis_id:
//     super_users:
//     cache_ttl:
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
//     - id:
//       num_counters:
//       max_cost:
//       buffer_items:
//   redis:
//     - id:
//       key_prefix:
//       with_inmem_id:
//       addr:
//       username:
//       password:
//       timeout:
//       max_retries:
//       min_retry_backoff:
//       max_retry_backoff:

const (
	DefaultSeqDBClientID     = "default_seq_db"
	DefaultCHClientID        = "default_ch"
	DefaultInmemCacheID      = "default_inmem_cache"
	DefaultRedisID           = "default_redis"
	DefaultRedis2ID          = "default_redis_2"
	DefaultMassExportRedisID = "mass_export"

	SeqDBClientModeGRPC = "grpc"

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

	defaultLogsLifespanCacheTTL = 10 * time.Minute

	defaultAdminCacheTTL = time.Hour

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

type TLS struct {
	RootCA     string `yaml:"root_ca"`
	CACert     string `yaml:"ca_cert"`
	PrivateKey string `yaml:"private_key"`
	Insecure   bool   `yaml:"insecure"`
}

type OIDC struct {
	CacheID        string   `yaml:"cache_id"`
	CacheSecretKey string   `yaml:"cache_secret_key"`
	SkipVerify     bool     `yaml:"skip_verify"`
	AuthURLs       []string `yaml:"auth_urls"`
	TLS            *TLS     `yaml:"tls"`
	AllowedClients []string `yaml:"allowed_clients"`
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
	KeyPrefix       string        `yaml:"key_prefix"`
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

func (c *Cache) InmemByID(id string) *InmemoryCache {
	if c == nil {
		return nil
	}

	for i := range c.Inmemory {
		if c.Inmemory[i].ID == id {
			return &c.Inmemory[i]
		}
	}

	return nil
}

func (c *Cache) RedisByID(id string) *Redis {
	if c == nil {
		return nil
	}

	for i := range c.Redis {
		if c.Redis[i].ID == id {
			return &c.Redis[i]
		}
	}

	return nil
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
	SeqDBID          string          `yaml:"seq_db_id"`
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

func (c *Clients) ClickHouseByID(id string) *CHClient {
	if c == nil {
		return nil
	}

	for i := range c.ClickHouse {
		if c.ClickHouse[i].ID == id {
			return &c.ClickHouse[i]
		}
	}

	return nil
}

type Handlers struct {
	SeqAPI      SeqAPI       `yaml:"seq_api"`
	ErrorGroups *ErrorGroups `yaml:"error_groups"`
	MassExport  *MassExport  `yaml:"mass_export"`
	Admin       *Admin       `yaml:"admin"`
	AsyncSearch AsyncSearch  `yaml:"async_search"`
}

type Admin struct {
	RedisID    string        `yaml:"redis_id"`
	SuperUsers []string      `yaml:"super_users"`
	CacheTTL   time.Duration `yaml:"cache_ttl"`
}

type Field struct {
	Name string `yaml:"name"`
	Type string `yaml:"type"`
}

type SeqAPI struct {
	SeqDBID       string               `yaml:"seq_db_id"`
	CacheID       string               `yaml:"cache_id"`
	RedisID       string               `yaml:"redis_id"`
	GlobalOptions SeqAPIGlobalOptions  `yaml:"global_options"`
	Options       SeqAPIOptions        `yaml:"options"`
	Envs          map[string]SeqAPIEnv `yaml:"envs"`
	DefaultEnv    string               `yaml:"default_env"`
}

type SeqAPIGlobalOptions struct {
	PinnedFields         []Field       `yaml:"pinned_fields"`
	SystemFields         []Field       `yaml:"system_fields"`
	Masking              *Masking      `yaml:"masking"`
	EventsCacheTTL       time.Duration `yaml:"events_cache_ttl"`
	LogsLifespanCacheTTL time.Duration `yaml:"logs_lifespan_cache_ttl"`
	FieldsCacheTTL       time.Duration `yaml:"fields_cache_ttl"`
}

type SeqAPIEnv struct {
	SeqDBID string         `yaml:"seq_db_id"`
	Options *SeqAPIOptions `yaml:"options"`
}

type SeqAPIOptions struct {
	MaxSearchLimit             int32 `yaml:"max_search_limit,omitempty"`
	MaxSearchTotalLimit        int64 `yaml:"max_search_total_limit,omitempty"`
	MaxSearchOffsetLimit       int32 `yaml:"max_search_offset_limit,omitempty"`
	MaxExportLimit             int32 `yaml:"max_export_limit,omitempty"`
	SeqCLIMaxSearchLimit       int   `yaml:"seq_cli_max_search_limit,omitempty"`
	MaxParallelExportRequests  int   `yaml:"max_parallel_export_requests,omitempty"`
	MaxAggregationsPerRequest  int   `yaml:"max_aggregations_per_request,omitempty"`
	MaxBucketsPerAggregationTs int   `yaml:"max_buckets_per_aggregation_ts,omitempty"`
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
	CHID           string                    `yaml:"ch_id"`
	Envs           map[string]ErrorGroupsEnv `yaml:"envs"`
	LogTagsMapping LogTagsMapping            `yaml:"log_tags_mapping"`
	QueryFilter    map[string]string         `yaml:"query_filter"`
	DefaultEnv     string                    `yaml:"default_env"`
}

type ErrorGroupsEnv struct {
	CHID string `yaml:"ch_id"`
}

type AsyncSearch struct {
	SeqDBID              string                    `yaml:"seq_db_id"`
	Envs                 map[string]AsyncSearchEnv `yaml:"envs"`
	AdminUsers           []string                  `yaml:"admin_users"`
	ListQueryLengthLimit int                       `yaml:"list_query_length_limit"`
	DefaultEnv           string                    `yaml:"default_env"`
}

type AsyncSearchEnv struct {
	SeqDBID              string `yaml:"seq_db_id"`
	ListQueryLengthLimit int    `yaml:"list_query_length_limit"`
}

func Normalize(cfg *Config) error {
	if len(cfg.Clients.SeqDB) == 0 {
		return fmt.Errorf("clients.seq_db must contain at least one client")
	}

	seqDBIDs := make(map[string]struct{}, len(cfg.Clients.SeqDB))
	for i := range cfg.Clients.SeqDB {
		c := &cfg.Clients.SeqDB[i]
		if c.ID == "" {
			return fmt.Errorf("clients.seq_db[%d].id cannot be empty", i)
		}
		if _, ok := seqDBIDs[c.ID]; ok {
			return fmt.Errorf("duplicate clients.seq_db.id %q", c.ID)
		}

		seqDBIDs[c.ID] = struct{}{}

		if c.ClientMode == "" {
			c.ClientMode = SeqDBClientModeGRPC
		} else if c.ClientMode != SeqDBClientModeGRPC {
			return fmt.Errorf("invalid clients.seq_db[%q].client_mode: %q (allowed: %q)", c.ID, c.ClientMode, SeqDBClientModeGRPC)
		}

		if c.GRPCKeepaliveParams != nil {
			c.GRPCKeepaliveParams.Time = max(c.GRPCKeepaliveParams.Time, minGRPCKeepaliveTime)
			c.GRPCKeepaliveParams.Timeout = max(c.GRPCKeepaliveParams.Timeout, minGRPCKeepaliveTimeout)
		}
	}

	chIDs := make(map[string]struct{}, len(cfg.Clients.ClickHouse))
	for i := range cfg.Clients.ClickHouse {
		ch := &cfg.Clients.ClickHouse[i]
		if ch.ID == "" {
			return fmt.Errorf("clients.clickhouse[%d].id cannot be empty", i)
		}
		if _, ok := chIDs[ch.ID]; ok {
			return fmt.Errorf("duplicate clients.clickhouse.id %q", ch.ID)
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
			return fmt.Errorf("cache.inmemory[%d].id cannot be empty", i)
		}
		if _, ok := inmemIDs[inm.ID]; ok {
			return fmt.Errorf("duplicate cache.inmemory.id %q", inm.ID)
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
			return fmt.Errorf("cache.redis[%d].id cannot be empty", i)
		}
		if _, ok := redisIDs[r.ID]; ok {
			return fmt.Errorf("duplicate cache.redis.id %q", r.ID)
		}

		redisIDs[r.ID] = struct{}{}

		if r.WithInmemID != "" {
			if _, ok := inmemIDs[r.WithInmemID]; !ok {
				return fmt.Errorf("cache.redis[%q].with_inmem_id %q not found in cache.inmemory", r.ID, r.WithInmemID)
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

	setSeqAPIGlobalOptionsDefaults(&cfg.Handlers.SeqAPI.GlobalOptions)
	setSeqAPIOptionsDefaults(&cfg.Handlers.SeqAPI.Options)

	if len(cfg.Handlers.SeqAPI.Envs) > 0 {
		if cfg.Handlers.SeqAPI.SeqDBID != "" {
			return fmt.Errorf("handlers.seq_api.seq_db_id must be empty when envs is used. Put seq_db_id inside each env")
		}
		if cfg.Handlers.SeqAPI.DefaultEnv == "" {
			return fmt.Errorf("handlers.seq_api.default_env must be specified when envs is used")
		}
		if _, exists := cfg.Handlers.SeqAPI.Envs[cfg.Handlers.SeqAPI.DefaultEnv]; !exists {
			return fmt.Errorf("handlers.seq_api.default_env %q not found in envs", cfg.Handlers.SeqAPI.DefaultEnv)
		}

		for envName, envConfig := range cfg.Handlers.SeqAPI.Envs {
			if envConfig.SeqDBID == "" {
				return fmt.Errorf("handlers.seq_api.envs[%q].seq_db_id cannot be empty", envName)
			}
			if _, ok := seqDBIDs[envConfig.SeqDBID]; !ok {
				return fmt.Errorf("unknown handlers.seq_api.envs[%q].seq_db_id %q", envName, envConfig.SeqDBID)
			}
		}
	} else {
		if cfg.Handlers.SeqAPI.SeqDBID == "" {
			return fmt.Errorf("handlers.seq_api.seq_db_id cannot be empty when envs is not used")
		}
		if _, ok := seqDBIDs[cfg.Handlers.SeqAPI.SeqDBID]; !ok {
			return fmt.Errorf("unknown handlers.seq_api.seq_db_id %q", cfg.Handlers.SeqAPI.SeqDBID)
		}
	}

	if cfg.Handlers.MassExport != nil {
		if cfg.Handlers.MassExport.SeqDBID == "" {
			return fmt.Errorf("handlers.mass_export.seq_db_id cannot be empty")
		}
		if _, ok := seqDBIDs[cfg.Handlers.MassExport.SeqDBID]; !ok {
			return fmt.Errorf("unknown handlers.mass_export.seq_db_id %q", cfg.Handlers.MassExport.SeqDBID)
		}

		if cfg.Handlers.MassExport.SessionStore != nil {
			if cfg.Handlers.MassExport.SessionStore.RedisID == "" {
				return fmt.Errorf("handlers.mass_export.session_store.redis_id cannot be empty")
			}
			if _, ok := redisIDs[cfg.Handlers.MassExport.SessionStore.RedisID]; !ok {
				return fmt.Errorf("unknown handlers.mass_export.session_store.redis_id %q", cfg.Handlers.MassExport.SessionStore.RedisID)
			}
		}
	}

	if len(cfg.Handlers.AsyncSearch.Envs) > 0 {
		if cfg.Handlers.AsyncSearch.SeqDBID != "" {
			return fmt.Errorf("handlers.async_search.seq_db_id must be empty when envs is used. Put seq_db_id inside each env")
		}
		if cfg.Handlers.AsyncSearch.DefaultEnv == "" {
			return fmt.Errorf("handlers.async_search.default_env must be specified when using envs")
		}
		if _, ok := cfg.Handlers.AsyncSearch.Envs[cfg.Handlers.AsyncSearch.DefaultEnv]; !ok {
			return fmt.Errorf("handlers.async_search.default_env %q not found in envs", cfg.Handlers.AsyncSearch.DefaultEnv)
		}

		for envName, envConfig := range cfg.Handlers.AsyncSearch.Envs {
			if envConfig.SeqDBID == "" {
				return fmt.Errorf("handlers.async_search.envs[%q].seq_db_id cannot be empty", envName)
			}
			if _, ok := seqDBIDs[envConfig.SeqDBID]; !ok {
				return fmt.Errorf("unknown handlers.async_search.envs[%q].seq_db_id %q", envName, envConfig.SeqDBID)
			}
			if envConfig.ListQueryLengthLimit <= 0 {
				envConfig.ListQueryLengthLimit = cfg.Handlers.AsyncSearch.ListQueryLengthLimit
			}

			cfg.Handlers.AsyncSearch.Envs[envName] = envConfig
		}
	} else {
		if cfg.Handlers.AsyncSearch.SeqDBID == "" {
			return fmt.Errorf("handlers.async_search.seq_db_id cannot be empty when envs is not used")
		}
		if _, ok := seqDBIDs[cfg.Handlers.AsyncSearch.SeqDBID]; !ok {
			return fmt.Errorf("unknown handlers.async_search.seq_db_id %q", cfg.Handlers.AsyncSearch.SeqDBID)
		}
	}

	if cfg.Handlers.ErrorGroups != nil {
		eg := cfg.Handlers.ErrorGroups
		if len(eg.Envs) > 0 {
			if eg.CHID != "" {
				return fmt.Errorf("handlers.error_groups.ch_id must be empty when envs is used. Put ch_id inside each env")
			}
			if eg.DefaultEnv == "" {
				return fmt.Errorf("handlers.error_groups.default_env must be specified when using envs")
			}
			if _, ok := eg.Envs[eg.DefaultEnv]; !ok {
				return fmt.Errorf("handlers.error_groups.default_env %q not found in envs", eg.DefaultEnv)
			}

			for envName, envConfig := range eg.Envs {
				if envConfig.CHID == "" {
					return fmt.Errorf("handlers.error_groups.envs[%q].ch_id cannot be empty", envName)
				}
				if _, ok := chIDs[envConfig.CHID]; !ok {
					return fmt.Errorf("unknown handlers.error_groups.envs[%q].ch_id %q", envName, envConfig.CHID)
				}
			}
		} else {
			if cfg.Handlers.ErrorGroups.CHID == "" {
				return fmt.Errorf("handlers.error_groups.ch_id cannot be empty when envs is not used")
			}
			if _, ok := chIDs[cfg.Handlers.ErrorGroups.CHID]; !ok {
				return fmt.Errorf("unknown handlers.error_groups.ch_id %q", cfg.Handlers.ErrorGroups.CHID)
			}
		}
	}

	if cfg.Server.Auth != nil && cfg.Server.Auth.OIDC != nil {
		oidc := cfg.Server.Auth.OIDC
		hasCacheKey := oidc.CacheSecretKey != ""
		hasCacheID := oidc.CacheID != ""

		switch {
		case hasCacheKey && !hasCacheID:
			return fmt.Errorf("auth.oidc.cache_secret_key is set but auth.oidc.cache_id is empty")
		case !hasCacheKey && hasCacheID:
			return fmt.Errorf("auth.oidc.cache_id is set but auth.oidc.cache_secret_key is empty")
		case hasCacheID && hasCacheKey:
			if cfg.Cache.InmemByID(oidc.CacheID) == nil && cfg.Cache.RedisByID(oidc.CacheID) == nil {
				return fmt.Errorf("auth.oidc.cache_id %q not found", oidc.CacheID)
			}
		}
	}

	if cfg.Handlers.Admin != nil {
		if cfg.Handlers.Admin.CacheTTL <= 0 {
			cfg.Handlers.Admin.CacheTTL = defaultAdminCacheTTL
		}

		if cfg.Handlers.Admin.RedisID != "" {
			redisCfg := cfg.Cache.RedisByID(cfg.Handlers.Admin.RedisID)
			if redisCfg == nil {
				return fmt.Errorf("unknown handlers.admin.redis_id %q", cfg.Handlers.Admin.RedisID)
			}
			if redisCfg.WithInmemID != "" {
				return fmt.Errorf("handlers.admin.redis_id %q: with_inmem_id is not allowed", cfg.Handlers.Admin.RedisID)
			}
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
}

func setSeqAPIGlobalOptionsDefaults(options *SeqAPIGlobalOptions) {
	if options.EventsCacheTTL <= 0 {
		options.EventsCacheTTL = defaultEventsCacheTTL
	}
	if options.LogsLifespanCacheTTL <= 0 {
		options.LogsLifespanCacheTTL = defaultLogsLifespanCacheTTL
	}
}
