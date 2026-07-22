package v1

import (
	"time"
)

// Deprecated configuration scheme
// version:
// server:
//   http_addr:
//   grpc_addr:
//   debug_addr:
//   grpc_connection_timeout:
//   http_read_timeout:
//   http_read_header_timeout:
//   http_write_timeout:
//   cors:
//     allowed_origins:
//     allowed_methods:
//     allowed_headers:
//     exposed_headers:
//     allow_credentials:
//     max_age:
//     options_passthrough:
//   jwt_secret_key:
//   oidc:
//     skip_verify:
//     auth_urls:
//     root_ca:
//     ca_cert:
//     private_key:
//     ssl_skip_verify:
//     allowed_clients:
//     cache_secret_key:
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
//   cache:
//     inmemory:
//       num_counters:
//       max_cost:
//       buffer_items:
//     redis:
//       addr:
//       username:
//       password:
//       timeout:
//       max_retries:
//       min_retry_backoff:
//       max_retry_backoff:
//   db:
//     name:
//     host:
//     port:
//     pass:
//     user:
//     request_timeout:
//     connection_pool_capacity:
//     use_prepared_statements:
//   clickhouse:
//     addrs:
//     database:
//     username:
//     password:
//     sharded:
//     dial_timeout:
//     read_timeout:
// clients:
//   seq_db_timeout:
//   seq_db_avg_doc_size:
//   seq_db_addrs:
//   request_retries:
//   initial_retry_backoff:
//   max_retry_backoff:
//   proxy_client_mode:
//   grpc_keepalive_params:
//     time:
//     timeout:
//     permit_without_stream:
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
// handlers:
//   seq_api:
//     max_search_limit:
//     max_search_total_limit:
//     max_search_offset_limit:
//     max_export_limit:
//     seq_cli_max_search_limit:
//     max_parallel_export_requests:
//     max_aggregations_per_request:
//     max_buckets_per_aggregation_ts:
//     events_cache_ttl:
//     pinned_fields:
//       name:
//       type:
//     system_fields:
//       name:
//       type:
//     logs_lifespan_cache_key:
//     logs_lifespan_cache_ttl:
//     fields_cache_ttl:
//     masking:
//       masks:
//         re:
//         groups:
//         mode:
//         replace_word:
//         process_fields:
//         ignore_fields:
//         field_filters:
//           condition:
//           filters:
//             field:
//             mode:
//             values:
//       process_fields:
//       ignore_fields:
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
//       redis:
//         addr:
//         username:
//         password:
//         timeout:
//         max_retries:
//         min_retry_backoff:
//         max_retry_backoff:
//       export_lifetime:
//     seq_proxy_downloader:
//       delay:
//       initial_retry_backoff:
//       max_retry_backoff:
//   async_search:
//     admin_users:
//     list_query_length_limit:

type Config struct {
	Version  *int      `yaml:"version"`
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

type Field struct {
	Name string `yaml:"name"`
	Type string `yaml:"type"`
}

type SeqAPI struct {
	*SeqAPIOptions `yaml:",inline"`
	Envs           map[string]SeqAPIEnv `yaml:"envs"`
	DefaultEnv     string               `yaml:"default_env"`
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
	Env     []string `yaml:"env"`
	Service []string `yaml:"service"`
	Release []string `yaml:"release"`
}

type ErrorGroups struct {
	LogTagsMapping LogTagsMapping    `yaml:"log_tags_mapping"`
	QueryFilter    map[string]string `yaml:"query_filter"`
}

type AsyncSearch struct {
	AdminUsers           []string `yaml:"admin_users"`
	ListQueryLengthLimit int      `yaml:"list_query_length_limit"`
}
