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
// 	   options:
// 	     oidc:
//         cache_id:
//         cache_secret_key:
//         skip_verify:
//         auth_urls:
//         tls:
//           root_ca:
//           ca_cert:
//           private_key:
//           insecure:
//         allowed_clients:
//       jwt:
//         secret_key:
// 	   envs:
// 	     <env_name>:
// 		   options: # same as root options
//   rate_limiters:
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
//       <env_name>:
//         options: # same as root options
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
//         cache_id:
//         redis_id:
//         ttl:
//           events:
//           logs_lifespan:
//           fields:
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
//     ch_id:
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
//     options:
//    	  admin_users:
//        list_query_length_limit:
//     envs:
//       <env_name>:
//         seq_db_id:
//         options: # same as root options
//     default_env:
//   admin:
//     options:
//       redis_id:
//       super_users:
//       cache_ttl:
//     envs:
//       <env_name>:
//         options: # same as root options
//     default_env:
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
//       with_inmem_id:
//       key_prefix:
//       addr:
//       username:
//       password:
//       timeout:
//       max_retries:
//       min_retry_backoff:
//       max_retry_backoff:

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
	WithInmemID     string        `yaml:"with_inmem_id"`
	KeyPrefix       string        `yaml:"key_prefix"`
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
	Options AuthOptions        `yaml:"options"`
	Envs    map[string]AuthEnv `yaml:"envs"`
}

type AuthEnv struct {
	Options *AuthOptions `yaml:"options"`
}

type AuthOptions struct {
	OIDC *OIDC `yaml:"oidc"`
	JWT  *JWT  `yaml:"jwt"`
}

type Server struct {
	HTTP         HTTP         `yaml:"http"`
	GRPC         GRPC         `yaml:"grpc"`
	Debug        Debug        `yaml:"debug"`
	Auth         *Auth        `yaml:"auth"`
	RateLimiters RateLimiters `yaml:"rate_limiters"`
}

type RateLimiters struct {
	Options ApiToRateLimiters          `yaml:"options"`
	Envs    map[string]RateLimitersEnv `yaml:"envs"`
}

type RateLimitersEnv struct {
	Options ApiToRateLimiters `yaml:"options"`
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
	MassExport  *MassExport  `yaml:"mass_export"`
	AsyncSearch *AsyncSearch `yaml:"async_search"`
	ErrorGroups *ErrorGroups `yaml:"error_groups"`
	Admin       *Admin       `yaml:"admin"`
}

type Admin struct {
	Options    AdminOptions        `yaml:"options"`
	Envs       map[string]AdminEnv `yaml:"envs"`
	DefaultEnv string              `yaml:"default_env"`
}

type AdminEnv struct {
	Options *AdminOptions `yaml:"options"`
}

type AdminOptions struct {
	RedisID    string        `yaml:"redis_id"`
	SuperUsers []string      `yaml:"super_users"`
	CacheTTL   time.Duration `yaml:"cache_ttl"`
}

type Field struct {
	Name string `yaml:"name"`
	Type string `yaml:"type"`
}

type SeqAPI struct {
	SeqDBID    string               `yaml:"seq_db_id"`
	Options    SeqAPIOptions        `yaml:"options"`
	Envs       map[string]SeqAPIEnv `yaml:"envs"`
	DefaultEnv string               `yaml:"default_env"`
}

type SeqAPIEnv struct {
	SeqDBID string         `yaml:"seq_db_id"`
	Options *SeqAPIOptions `yaml:"options"`
}

type SeqAPIOptions struct {
	Limits       SeqAPILimits `yaml:"limits"`
	Caches       SeqAPICaches `yaml:"caches"`
	Masking      *Masking     `yaml:"masking"`
	PinnedFields []Field      `yaml:"pinned_fields"`
	SystemFields []Field      `yaml:"system_fields"`
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
	CacheID string          `yaml:"cache_id"` // redis + inmem
	RedisID string          `yaml:"redis_id"` // redis only
	TTL     SeqAPICachesTTL `yaml:"ttl"`
}

type SeqAPICachesTTL struct {
	Events       time.Duration `yaml:"events"`
	LogsLifespan time.Duration `yaml:"logs_lifespan"`
	Fields       time.Duration `yaml:"fields"`
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
	CHID       string                    `yaml:"ch_id"`
	Options    ErrorGroupsOptions        `yaml:"options"`
	Envs       map[string]ErrorGroupsEnv `yaml:"envs"`
	DefaultEnv string                    `yaml:"default_env"`
}

type ErrorGroupsEnv struct {
	CHID    string              `yaml:"ch_id"`
	Options *ErrorGroupsOptions `yaml:"options"`
}

type ErrorGroupsOptions struct {
	LogTagsMapping LogTagsMapping    `yaml:"log_tags_mapping"`
	QueryFilter    map[string]string `yaml:"query_filter"`
}

type AsyncSearch struct {
	SeqDBID    string                    `yaml:"seq_db_id"`
	Options    AsyncSearchOptions        `yaml:"options"`
	Envs       map[string]AsyncSearchEnv `yaml:"envs"`
	DefaultEnv string                    `yaml:"default_env"`
}

type AsyncSearchEnv struct {
	SeqDBID string              `yaml:"seq_db_id"`
	Options *AsyncSearchOptions `yaml:"options"`
}

type AsyncSearchOptions struct {
	AdminUsers           []string `yaml:"admin_users"`
	ListQueryLengthLimit int      `yaml:"list_query_length_limit"`
}
