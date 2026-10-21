package config

import (
	"fmt"
	"slices"
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
// 	   read_timeout:
// 	   read_header_timeout:
//     write_timeout:
//     cors:
//       allowed_origins:
//       allowed_methods:
//       allowed_headers:
//       exposed_headers:
//       allow_credentials:
//       max_age:
//       options_passthrough:
//   auth:
//     options:
//       oidc:
//         cache_id:
//         cache_secret_key:
//         skip_verify:
//         auth_urls:
//         tls:
//           ca_cert:
//           client_cert:
//           client_key:
//           insecure:
//         allowed_clients:
//       jwt:
//         secret_key:
//     envs:
//       <env_name>: # same as root options
//   rate_limiting:
//     options:
//       <api_name>:
//         default:
//           rate_per_sec:
//           max_burst:
//           store_max_keys:
//           per_handler:
//         spec_users:
//           <username>:
//             rate_per_sec:
//             max_burst:
//             store_max_keys:
//             per_handler:
//     envs:
//       <env_name>: # same as root options
// clients:
//   seq_db:
//     - id:
//       timeout:
//       avg_doc_size:
//       addrs:
//       request_retries_backoff:
//       backoff:
// 		   initial_interval:
// 		   max_interval
//       client_mode:
//       grpc_params:
//         keepalive:
//           time:
//           timeout:
//           permit_without_stream:
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
//     options:
//       limits:
//         max_search_limit:
//         max_search_total:
//         max_search_offset:
//         max_export_limit:
//         seq_cli_max_search_limit:
//         max_parallel_export_requests:
//         max_aggregations_per_request:
//         max_buckets_per_aggregation_ts:
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
//       caches:
//         events:
//           id:
//           ttl:
//         logs_lifespan:
//           id:
//           ttl:
//         fields:
//           id:
//           ttl:
//       pinned_fields:
//         - name:
//           type:
//       system_fields:
//         - name:
//           type:
//     envs:
//       <env_name>:
//         seq_db_id:
//         options: # same as root options
//     default_env:
//   error_groups:
//     options:
//       log_tags_mapping:
//         env:
//         service:
//         release:
//       query_filter:
//         <ch_column>:
//     envs:
//       <env_name>:
//         ch_id:
//         options: # same as root options
//     default_env:
//   mass_export:
//     options:
//       workers_count:
//       tasks_channel_size:
//       part_length:
//       url_prefix:
//       allowed_users:
//       file_store:
//         s3:
//           endpoint:
//           access_key_id:
//           secret_access_key:
//           bucket_name:
//           enable_ssl:
//       session_store:
//         lifetime:
//         cache_id:
//       download_params:
//         batch_size:
//         delay:
// 		   delay_backoff:
//           initial_interval:
//           max_interval:
//     envs:
//       <env_name>:
//         seq_db_id:
//         options: # same as root options
//     default_env:
//   async_search:
//     options:
//       admin_users:
//       list_query_length_limit:
//     envs:
//       <env_name>:
//         seq_db_id:
//         options: # same as root options
//     default_env:
//   admin:
//     options:
//       super_users:
//       cache:
//         id:
//         ttl:
//     envs:
//       <env_name>:
//         options: # same as root options
//     default_env:
// db:
//   name:
//   host:
//   port:
//   user:
//   pass:
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
//       with_inmem_id:
//       key_prefix:
//       addr:
//       username:
//       password:
//       max_retries:
//       timeout:
//       min_retry_backoff:
//       max_retry_backoff:

type Config struct {
	Version  int      `yaml:"version"`
	Server   Server   `yaml:"server"`
	Clients  Clients  `yaml:"clients"`
	Handlers Handlers `yaml:"handlers"`
	Cache    Cache    `yaml:"cache,omitempty"`
	DB       *DB      `yaml:"db,omitempty"`
}

type CORS struct {
	AllowedOrigins     []string `yaml:"allowed_origins,omitempty"`
	AllowedMethods     []string `yaml:"allowed_methods,omitempty"`
	AllowedHeaders     []string `yaml:"allowed_headers,omitempty"`
	ExposedHeaders     []string `yaml:"exposed_headers,omitempty"`
	MaxAge             int      `yaml:"max_age,omitempty"`
	AllowCredentials   bool     `yaml:"allow_credentials,omitempty"`
	OptionsPassthrough bool     `yaml:"options_passthrough,omitempty"`
}

type TLS struct {
	CACert     string `yaml:"ca_cert,omitempty"`
	ClientCert string `yaml:"client_cert,omitempty"`
	ClientKey  string `yaml:"client_key,omitempty"`
	Insecure   bool   `yaml:"insecure,omitempty"`
}

type OIDC struct {
	CacheID        string   `yaml:"cache_id,omitempty"`
	CacheSecretKey string   `yaml:"cache_secret_key,omitempty"`
	SkipVerify     bool     `yaml:"skip_verify,omitempty"`
	AuthURLs       []string `yaml:"auth_urls"`
	AllowedClients []string `yaml:"allowed_clients,omitempty"`
	TLS            *TLS     `yaml:"tls,omitempty"`
}

type DB struct {
	Name                   string        `yaml:"name"`
	Host                   string        `yaml:"host"`
	Port                   int64         `yaml:"port"`
	User                   string        `yaml:"user"`
	Pass                   string        `yaml:"pass"`
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
		MaxBurst     int  `yaml:"max_burst,omitempty"`
		StoreMaxKeys int  `yaml:"store_max_keys,omitempty"`
		PerHandler   bool `yaml:"per_handler,omitempty"`
	}

	RateLimiters struct {
		Default      RateLimiter            `yaml:"default"`
		SpecialUsers map[string]RateLimiter `yaml:"spec_users,omitempty"`
	}

	ApiToRateLimiters map[string]RateLimiters
)

type Inmemory struct {
	ID          string `yaml:"id"`
	NumCounters int64  `yaml:"num_counters,omitempty"`
	MaxCost     int64  `yaml:"max_cost,omitempty"`
	BufferItems int64  `yaml:"buffer_items,omitempty"`
}

type Redis struct {
	ID              string        `yaml:"id"`
	WithInmemID     string        `yaml:"with_inmem_id,omitempty"`
	KeyPrefix       string        `yaml:"key_prefix,omitempty"`
	Addr            string        `yaml:"addr"`
	Username        string        `yaml:"username,omitempty"`
	Password        string        `yaml:"password,omitempty"`
	MaxRetries      int           `yaml:"max_retries,omitempty"`
	Timeout         time.Duration `yaml:"timeout,omitempty"`
	MinRetryBackoff time.Duration `yaml:"min_retry_backoff,omitempty"`
	MaxRetryBackoff time.Duration `yaml:"max_retry_backoff,omitempty"`
}

type Cache struct {
	Inmemory []Inmemory `yaml:"inmemory,omitempty"`
	Redis    []Redis    `yaml:"redis,omitempty"`
}

func (c *Cache) InmemByID(id string) *Inmemory {
	i := slices.IndexFunc(c.Inmemory, func(inm Inmemory) bool {
		return inm.ID == id
	})
	if i == -1 {
		return nil
	}

	return &c.Inmemory[i]
}

func (c *Cache) RedisByID(id string) *Redis {
	i := slices.IndexFunc(c.Redis, func(r Redis) bool {
		return r.ID == id
	})
	if i == -1 {
		return nil
	}

	return &c.Redis[i]
}

type S3 struct {
	Endpoint        string `yaml:"endpoint"`
	AccessKeyID     string `yaml:"access_key_id"`
	SecretAccessKey string `yaml:"secret_access_key"`
	BucketName      string `yaml:"bucket_name"`
	EnableSSL       bool   `yaml:"enable_ssl,omitempty"`
}

type DownloadParams struct {
	BatchSize    uint64        `yaml:"batch_size,omitempty"`
	Delay        time.Duration `yaml:"delay,omitempty"`
	DelayBackoff Backoff       `yaml:"delay_backoff,omitempty"`
}

type SessionStore struct {
	RedisCacheID string        `yaml:"redis_cache_id"`
	Lifetime     time.Duration `yaml:"lifetime,omitempty"`
}

type FileStore struct {
	S3 *S3 `yaml:"s3"`
}

type MassExport struct {
	Options    MassExportOptions        `yaml:"options"`
	Envs       map[string]MassExportEnv `yaml:"envs,omitempty"`
	DefaultEnv string                   `yaml:"default_env,omitempty"`
}

type MassExportEnv struct {
	SeqDBID string             `yaml:"seq_db_id"`
	Options *MassExportOptions `yaml:"options,omitempty"`
}

type MassExportOptions struct {
	WorkersCount     int            `yaml:"workers_count"`
	TasksChannelSize int            `yaml:"tasks_channel_size,omitempty"`
	PartLength       time.Duration  `yaml:"part_length,omitempty"`
	URLPrefix        string         `yaml:"url_prefix"`
	AllowedUsers     []string       `yaml:"allowed_users,omitempty"`
	FileStore        FileStore      `yaml:"file_store"`
	SessionStore     SessionStore   `yaml:"session_store"`
	DownloadParams   DownloadParams `yaml:"download_params,omitempty"`
}

type HTTP struct {
	Addr              string        `yaml:"addr"`
	ReadHeaderTimeout time.Duration `yaml:"read_header_timeout,omitempty"`
	ReadTimeout       time.Duration `yaml:"read_timeout,omitempty"`
	WriteTimeout      time.Duration `yaml:"write_timeout,omitempty"`
	CORS              *CORS         `yaml:"cors,omitempty"`
}

type GRPC struct {
	Addr              string        `yaml:"addr"`
	ConnectionTimeout time.Duration `yaml:"connection_timeout,omitempty"`
}

type Debug struct {
	Addr              string        `yaml:"addr"`
	ReadHeaderTimeout time.Duration `yaml:"read_header_timeout,omitempty"`
	ReadTimeout       time.Duration `yaml:"read_timeout,omitempty"`
	WriteTimeout      time.Duration `yaml:"write_timeout,omitempty"`
	CORS              *CORS         `yaml:"cors,omitempty"`
}

type JWT struct {
	SecretKey string `yaml:"secret_key"`
}

type Auth struct {
	Options AuthOptions            `yaml:"options"`
	Envs    map[string]AuthOptions `yaml:"envs,omitempty"`
}

type AuthOptions struct {
	OIDC *OIDC `yaml:"oidc,omitempty"`
	JWT  *JWT  `yaml:"jwt,omitempty"`
}

type Server struct {
	HTTP         *HTTP         `yaml:"http,omitempty"`
	GRPC         *GRPC         `yaml:"grpc,omitempty"`
	Debug        Debug         `yaml:"debug"`
	Auth         *Auth         `yaml:"auth,omitempty"`
	RateLimiting *RateLimiting `yaml:"rate_limiting,omitempty"`
}

type RateLimiting struct {
	// rl options are optional, without them requests not covered by envs are not limited.
	Options ApiToRateLimiters            `yaml:"options,omitempty"`
	Envs    map[string]ApiToRateLimiters `yaml:"envs,omitempty"`
}

type KeepaliveParams struct {
	// After a duration of this time if the client doesn't see any activity it
	// pings the server to see if the transport is still alive.
	// If set below 10s, a minimum value of 10s will be used instead.
	Time time.Duration `yaml:"time,omitempty"`
	// After having pinged for keepalive check, the client waits for a duration
	// of Timeout and if no activity is seen even after that the connection is
	// closed. If set below 1s, a minimum value of 1s will be used instead.
	Timeout time.Duration `yaml:"timeout,omitempty"`
	// If true, client sends keepalive pings even with no active RPCs. If false,
	// when there are no active RPCs, Time and Timeout will be ignored and no
	// keepalive pings will be sent. False by default.
	PermitWithoutStream bool `yaml:"permit_without_stream,omitempty"`
}

type GRPCParams struct {
	Keepalive *KeepaliveParams `yaml:"keepalive,omitempty"`
}

type SeqDBClient struct {
	ID                    string        `yaml:"id"`
	Timeout               time.Duration `yaml:"timeout"`
	AvgDocSize            int           `yaml:"avg_doc_size,omitempty"`
	Addrs                 []string      `yaml:"addrs"`
	RequestRetries        int           `yaml:"request_retries,omitempty"`
	RequestRetriesBackoff *Backoff      `yaml:"request_retries_backoff,omitempty"`
	ClientMode            string        `yaml:"client_mode,omitempty"`
	GRPCParams            *GRPCParams   `yaml:"grpc_params,omitempty"`
}

type Backoff struct {
	InitialInterval time.Duration `yaml:"initial_interval,omitempty"`
	MaxInterval     time.Duration `yaml:"max_interval,omitempty"`
}

type CHClient struct {
	ID          string        `yaml:"id"`
	Addrs       []string      `yaml:"addrs"`
	Database    string        `yaml:"database,omitempty"`
	Username    string        `yaml:"username,omitempty"`
	Password    string        `yaml:"password,omitempty"`
	Sharded     bool          `yaml:"sharded,omitempty"`
	DialTimeout time.Duration `yaml:"dial_timeout,omitempty"`
	ReadTimeout time.Duration `yaml:"read_timeout,omitempty"`
}

type Clients struct {
	SeqDB      []SeqDBClient `yaml:"seq_db"`
	ClickHouse []CHClient    `yaml:"clickhouse,omitempty"`
}

func (c *Clients) ClickHouseByID(id string) *CHClient {
	i := slices.IndexFunc(c.ClickHouse, func(ch CHClient) bool {
		return ch.ID == id
	})
	if i == -1 {
		return nil
	}

	return &c.ClickHouse[i]
}

type Handlers struct {
	SeqAPI      SeqAPI       `yaml:"seq_api"`
	MassExport  *MassExport  `yaml:"mass_export,omitempty"`
	AsyncSearch *AsyncSearch `yaml:"async_search,omitempty"`
	ErrorGroups *ErrorGroups `yaml:"error_groups,omitempty"`
	Admin       *Admin       `yaml:"admin,omitempty"`
}

type Admin struct {
	Options    AdminOptions        `yaml:"options"`
	Envs       map[string]AdminEnv `yaml:"envs,omitempty"`
	DefaultEnv string              `yaml:"default_env,omitempty"`
}

type AdminEnv struct {
	Options *AdminOptions `yaml:"options,omitempty"`
}

type AdminOptions struct {
	SuperUsers []string      `yaml:"super_users,omitempty"`
	Cache      *HandlerCache `yaml:"cache,omitempty"`
}

type Field struct {
	Name string `yaml:"name"`
	Type string `yaml:"type"`
}

type SeqAPI struct {
	Options    SeqAPIOptions        `yaml:"options"`
	Envs       map[string]SeqAPIEnv `yaml:"envs,omitempty"`
	DefaultEnv string               `yaml:"default_env,omitempty"`
}

type SeqAPIEnv struct {
	SeqDBID string         `yaml:"seq_db_id"`
	Options *SeqAPIOptions `yaml:"options,omitempty"`
}

type SeqAPIOptions struct {
	Limits       SeqAPILimits `yaml:"limits,omitempty"`
	Caches       SeqAPICaches `yaml:"caches,omitempty"`
	Masking      *Masking     `yaml:"masking,omitempty"`
	PinnedFields []Field      `yaml:"pinned_fields,omitempty"`
	SystemFields []Field      `yaml:"system_fields,omitempty"`
}

type SeqAPILimits struct {
	MaxSearchLimit             int32 `yaml:"max_search_limit,omitempty"`
	MaxSearchTotal             int64 `yaml:"max_search_total,omitempty"`
	MaxSearchOffset            int32 `yaml:"max_search_offset,omitempty"`
	MaxExportLimit             int32 `yaml:"max_export_limit,omitempty"`
	SeqCLIMaxSearchLimit       int   `yaml:"seq_cli_max_search_limit,omitempty"`
	MaxParallelExportRequests  int   `yaml:"max_parallel_export_requests,omitempty"`
	MaxAggregationsPerRequest  int   `yaml:"max_aggregations_per_request,omitempty"`
	MaxBucketsPerAggregationTs int   `yaml:"max_buckets_per_aggregation_ts,omitempty"`
}

type SeqAPICaches struct {
	Events       *HandlerCache `yaml:"events,omitempty"`
	LogsLifespan *HandlerCache `yaml:"logs_lifespan,omitempty"`
	Fields       *HandlerCache `yaml:"fields,omitempty"`
}

type HandlerCache struct {
	ID  string        `yaml:"id"`
	TTL time.Duration `yaml:"ttl,omitempty"`
}

type Masking struct {
	Masks         []Mask   `yaml:"masks"`
	ProcessFields []string `yaml:"process_fields,omitempty"`
	IgnoreFields  []string `yaml:"ignore_fields,omitempty"`
}

type Mask struct {
	Re          string `yaml:"re"`
	Groups      []int  `yaml:"groups,omitempty"`
	Mode        string `yaml:"mode"`                   // "mask" or "replace" or "cut"
	ReplaceWord string `yaml:"replace_word,omitempty"` // for mode:replace

	ProcessFields []string `yaml:"process_fields,omitempty"`
	IgnoreFields  []string `yaml:"ignore_fields,omitempty"`

	FieldFilters *FieldFilterSet `yaml:"field_filters,omitempty"`
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
	Release []string `yaml:"release,omitempty"`
	Service []string `yaml:"service,omitempty"`
	Env     []string `yaml:"env,omitempty"`
}

type ErrorGroups struct {
	Options    ErrorGroupsOptions        `yaml:"options"`
	Envs       map[string]ErrorGroupsEnv `yaml:"envs,omitempty"`
	DefaultEnv string                    `yaml:"default_env,omitempty"`
}

type ErrorGroupsEnv struct {
	CHID    string              `yaml:"ch_id"`
	Options *ErrorGroupsOptions `yaml:"options,omitempty"`
}

type ErrorGroupsOptions struct {
	LogTagsMapping LogTagsMapping    `yaml:"log_tags_mapping"`
	QueryFilter    map[string]string `yaml:"query_filter,omitempty"`
}

type AsyncSearch struct {
	Options    AsyncSearchOptions        `yaml:"options"`
	Envs       map[string]AsyncSearchEnv `yaml:"envs,omitempty"`
	DefaultEnv string                    `yaml:"default_env,omitempty"`
}

type AsyncSearchEnv struct {
	SeqDBID string              `yaml:"seq_db_id"`
	Options *AsyncSearchOptions `yaml:"options,omitempty"`
}

type AsyncSearchOptions struct {
	AdminUsers           []string `yaml:"admin_users,omitempty"`
	ListQueryLengthLimit int      `yaml:"list_query_length_limit,omitempty"`
}
