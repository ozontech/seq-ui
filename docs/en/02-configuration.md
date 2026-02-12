# Configuration

> Despite the fact that there are a huge number of parameters in the configuration, not all of them are supported by UI at the moment. As the UI evolves, more and more of the parameters will be relevant.

The configuration is set via a `yaml`-file and environment variables.
`yaml`-file consists of three sections:
- [server](#server) - seq-ui server configutarion
- [clients](#clients) - seq-db clients configuration
- [handlers](#handlers) - seq-ui api handlers configuration

Sections configurable via environment variables:
- [tracing](#tracing) - seq-ui tracing configuration

You can specify your config file when running seq-ui by providing it with flag `--config`:
```shell
go run ./cmd/seq-ui --config <path-to-config-file>
```

## Config examples

See config examples in `config` directory:
- [example](https://github.com/ozontech/seq-ui/tree/main/config/config.example.yaml) - the minimal config

## Server

```yaml
server:
  http_addr:
  grpc_addr:
  debug_addr:
  grpc_connection_timeout:
  http_read_timeout:
  http_read_header_timeout:
  http_write_timeout:
  cors:
  jwt_secret_key:
  oidc:
  rate_limiters:
  cache:
  db:
  clickhouse:
```

### gRPC/HTTP server

**`http_addr`** *`string`* *`required`*

Host for HTTP server.

**`grpc_addr`** *`string`* *`required`*

Host for gRPC server.

**`debug_addr`** *`string`* *`required`*

Host for debug server.

**`grpc_connection_timeout`** *`string`* *`default="0"`*

The timeout for connection establishment (up to and including HTTP/2 handshaking) for all new connections in gRPC server. A zero or negative value will result in an immediate timeout.

> The value must be passed in the duration format: `<number>(ms|s|m|h)`.

**`http_read_timeout`** *`string`* *`default="0"`*

The maximum duration for reading the entire request, including the body. A zero or negative value means there will be no timeout.

> The value must be passed in the duration format: `<number>(ms|s|m|h)`.

**`http_read_header_timeout`** *`string`* *`default="0"`*

The amount of time allowed to read request headers. If zero, the value of `http_read_timeout` is used. If both are zero or negative value, there is no timeout.

> The value must be passed in the duration format: `<number>(ms|s|m|h)`.

**`http_write_timeout`** *`string`* *`default="0"`*

The maximum duration after header is read and before timing out writes of the response. A zero or negative value means there will be no timeout.

> The value must be passed in the duration format: `<number>(ms|s|m|h)`.

**`cors`** *`CORS`* *`optional`*

HTTP server CORS config. If not set, no CORS settings will be applied.

`CORS` fields:

+ **`allowed_origins`** *`[]string`* *`default=["*"]`*

  A list of origins a cross-domain request can be executed from. If the special `"*"` value is present in the list, all origins will be allowed.

+ **`allowed_methods`** *`[]string`* *`default=["HEAD", "GET", "POST", "PATCH", "DELETE"]`*

  A list of methods the client is allowed to use with cross-domain requests.

+ **`allowed_headers`** *`[]string`* *`default=[]`*

  A list of non simple headers the client is allowed to use with cross-domain requests. If the special `"*"` value is present in the list, all headers will be allowed.
  
  > `"Origin"` header is always appended to the list.

+ **`exposed_headers`** *`[]string`* *`default=[]`*

  Indicates which headers are safe to expose to the API of a CORS API specification.

+ **`allow_credentials`** *`bool`* *`default=false`*

  Indicates whether the request can include user credentials like cookies, HTTP authentication or client side SSL certificates.

+ **`max_age`** *`int`* *`default=0`*

  Indicates how long (in seconds) the results of a preflight request can be cached.

+ **`options_passthrough`** *`bool`* *`default=false`*

  Instructs preflight to let other potential next handlers to process the OPTIONS method. Turn this on if you handles `OPTIONS`.

### Auth

**`jwt_secret_key`** *`string`*  *`default=""`*

If set to non-empty string, JWT provider is created for API tokens verification.

> API tokens allow access only for [`/seqapi/*`, `/massexport/*`] routes in HTTP API and [`SeqAPIService`, `MassExportService`] service in gRPC API. For other routes/services requiring auth, OIDC check will be performed, so `jwt_secret_key` should be used in pair with `oidc`.

**`oidc`** *`OIDC`* *`optional`*

Open ID Connect config. If not set, no OIDC verification will be applied.

`OIDC` fields:

+ **`auth_urls`** *`[]string`* *`required`*

  List of OIDC auth URLs for sending verification requests. For each verification, the entire `auth_urls` list will be searched, choosing a URL.
  
+ **`root_ca`** *`string`* *`default=""`*

  Path to file with CA root certificate or the certificate itself. If set, it will be passed to OIDC client tls config.

+ **`ca_cert`** *`string`* *`default=""`*

  Path to file with CA certificate or the certificate itself. If set, it will be passed to OIDC client tls config.

+ **`private_key`** *`string`* *`default=""`*

  Path to file with private key generated with CA certificate or the private key itself. If set, it will be passed to OIDC client tls config.

+ **`ssl_skip_verify`** *`bool`* *`default=false`*

  If set, disables security checks on OIDC client.

+ **`allowed_clients`** *`[]string`*  *`default=[]`*

  List of allowed clients. If set, only the specified clients will be verified. The `Audience` token field is used.

+ **`skip_verify`** *`bool`* *`default=false`*

  If set, only the issuer and expiration are checked locally without requests to `auth_urls`.

+ **`cache_secret_key`** *`string`* *`default=""`*

  If set to non-empty string, OIDC tokens are cached using `cache_secret_key` until the token expiration.

### Rate limiting

**`rate_limiters`** *`map[string]ApiRateLimiters`* *`optional`*

gRPC and HTTP server rate limiters configs. If not set, rate limiting will not be applied.

> Map key is HTTP API base route or gRPC API service name:
> - `seqapi`, `userprofile`, `dashboards`, `massexport`, `errorgroups`
> - `SeqAPIService`, `UserProfileService`, `DashboardsService`, `MassExportService`, `ErrorGroupsService`

`ApiRateLimiters` fields:

+ **`default`** *`RateLimiter`* *`required`*

  Describes default rate limiter for unspecified and unauthorized users.

+ **`spec_users`** *`map[string]RateLimiter`* *`optional`*

  Describes rate limiters for special users and tokens. Key is username or token name.

`RateLimiter` fields:

+ **`rate_per_sec`** *`int`* *`required`*

  Number of requests allowed per second. If [auth](#auth) is active, each user gets personal quota, otherwise the quota is general for all requests.

  > The app utilizes rate limiter based on [Generic Cell Rate Algorithm (GCRA)](https://en.wikipedia.org/wiki/Generic_cell_rate_algorithm). If `max_burst` is set to zero, the next request will be allowed after `(1 sec / rate_per_sec)` time (e.g. `rate_per_sec` is set to 2, then the next request will be allowed after 500ms).

+ **`max_burst`** *`int`* *`default=0`*

  Number of requests that will be allowed to exceed the rate in a single burst.

  > Total amount of requests allowed per second can be higher than `rate_per_sec`, if `max_burst` is greater than zero.

+ **`store_max_keys`** *`int`* *`default=0`*

  Max amount of keys to be stored in rate limiter store. A zero or negative value means that the amount of keys is considered unlimited.

+ **`per_handler`** *`bool`* *`default=false`*

  If set, every API handler will be limited separately.

### Cache

**`cache`** *`Cache`* *`optional`*

Cache config.

`Cache` fields:

+ **`inmemory`** *`InmemoryCache`* *`optional`*

  Config for in-memory cache.

  > If not set, it will be applied with default values.

+ **`redis`** *`Redis`* *`optional`*

  Config for redis cache.

  It works in pair with in-memory cache:
  + when a key-value pair sets in redis, it's also sets in in-memory cache
  + when in-memory cache contains the key, the request to redis doesn't occur
  + when in-memory cache doesn't contain the key, but redis contains, the redis result sets in in-memory cache

  > If redis isn't available, it falls back to in-memory cache.

`InmemoryCache` fields:

+ **`num_counters`** *`int`* *`default=1e7`*

  The number of counters (keys) to keep that hold access frequency information. If set to zero or negative value, then it will be reset to `default`.

  > It's generally a good idea to have more counters than the max cache capacity (`max_cost`), as this will improve eviction accuracy and subsequent hit ratios.

+ **`max_cost`** *`int`* *`default=1e6`*

  Cache capacity. If set to zero or negative value, then it will be reset to `default`.

+ **`buffer_items`** *`int`* *`default=64`*

  Determines the size of Get buffers. If set to zero or negative value, then it will be reset to `default`.

`Redis` fields:

+ **`addr`** *`string`* *`required`*

  Address in `host:port` format.

+ **`username`** *`string`* *`default=""`*

  Username to authenticate the connection when connecting to a Redis 6.0 instance, or greater, that is using the Redis ACL system.

+ **`password`** *`string`* *`default=""`*

  Password specified in the `requirepass` server configuration option (if connecting to a Redis 5.0 instance, or lower), or the User Password when connecting to a Redis 6.0 instance, or greater, that is using the Redis ACL system.

+ **`timeout`** *`string`* *`default="3s"`*

  Read/write timeout. If set to `-1`, disables timeout.

  > The value must be passed in the duration format: `<number>(ms|s|m|h)`.

+ **`max_retries`** *`int`* *`default=3`*

  Maximum number of retries before giving up. If set to `-1`, disables retries.

+ **`min_retry_backoff`** *`string`* *`default="8ms"`*

  Minimum backoff between each retry. If set to `-1`, disables backoff.

  > The value must be passed in the duration format: `<number>(ms|s|m|h)`.

+ **`max_retry_backoff`** *`string`* *`default="512ms"`*

  Maximum backoff between each retry. If set to `-1`, disables backoff.

  > The value must be passed in the duration format: `<number>(ms|s|m|h)`.

### External storages

**`db`** *`DB`* *`optional`*

PostgreSQL database config.

> Required for `/userprofile` and `seqapi/v1/async_search/` handlers.

`DB` fields:

+ **`name`** *`string`* *`required`*

  Database name.

+ **`host`** *`string`* *`required`*

  Database host.

+ **`port`** *`int`* *`required`*

  Database port.

+ **`user`** *`string`* *`required`*

  Database username.

+ **`pass`** *`string`* *`required`*

  Database password.

+ **`connection_pool_capacity`** *`int`* *`required`*

  The maximum connection pool size.

+ **`request_timeout`** *`string`* *`required`*

  Timeout for all database requests.

  > The value must be passed in the duration format: `<number>(ms|s|m|h)`.

+ **`use_prepared_statements`** *`bool`* *`default=true`*

  If set to `false`, disables the use of postgres prepared statements.

**`clickhouse`** *`ClickHouse`* *`optional`*

> Required for `/errorgroups` handlers.

ClickHouse database config.

`ClickHouse` fields:

+ **`database`** *`string`* *`required`*

  Database name.

+ **`username`** *`string`* *`required`*

  Database username.

+ **`password`** *`int`* *`required`*

  Database password.

+ **`dial_timeout`** *`int`* *`default="5s"`*

  Database dial timeout. If set to zero or negative value, then it will be reset to `default`.

  > The value must be passed in the duration format: `<number>(ms|s|m|h)`.

+ **`read_timeout`** *`int`* *`default="30s"`*

  Database read timeout. If set to zero or negative value, then it will be reset to `default`.

  > The value must be passed in the duration format: `<number>(ms|s|m|h)`.

+ **`sharded`** *`bool`* *`default=false`*

  Indicates whether the clickhouse is sharded. It is used for some queries that depend on the clickhouse scheme.

## Clients

```yaml
clients:
  seq_db_addrs:
  proxy_client_mode:
  seq_db_timeout:
  seq_db_avg_doc_size:
  request_retries:
  initial_retry_backoff:
  max_retry_backoff:
  grpc_keepalive_params:
  seq_db:
```

**`seq_db_addrs`** *`[]string`* *`required`*

List of seq-db proxy hosts to be used in client calls. If there are more than one host, for each request random host will be chosen.

> Deprecated. Use `seq_db.addrs` instead.

**`proxy_client_mode`** *`string`* *`default="grpc"`* *`options="grpc"`*

Allow choosing how to send requests to seq-db.

> Deprecated. Use `seq_db.client_mode` instead.

**`seq_db_timeout`** *`string`* *`default="0"`*

Timeout for requests made by the client. A zero value means no timeout.

> The value must be passed in the duration format: `<number>(ms|s|m|h)`.

> Deprecated. Use `seq_db.timeout` instead.

**`seq_db_avg_doc_size`** *`int`* *`default=0`*

Specifies the average documents size in `KB` that the client calls returns. It's used in combination with `handlers.seq_api.max_search_limit` to calculate the maximum response size per client request.

> Regardless of `seq_db_avg_doc_size`, the minimum response size per client request is `4MB`.

> Deprecated. Use `seq_db.avg_doc_size` instead.

**`request_retries`** *`int`* *`default=0`*

The number of retries to send a request to client after the first attempt. For each retry, the entire `seq_db_addrs` list will be searched, choosing a random host. A zero value means no retries. If set to negative value, then it will be reset to `default`.

> Deprecated. Use `seq_db.request_retries` instead.

**`initial_retry_backoff`** *`string`* *`default="0"`*

Initial duration interval value to be used in backoff with retries. If set to `0`, disables backoff.

> The value must be passed in the duration format: `<number>(ms|s|m|h)`.

> Deprecated. Use `seq_db.initial_retry_backoff` instead.

**`max_retry_backoff`** *`string`* *`default="0"`*

Max duration interval value to be used in backoff with retries. If set to `0`, only value from `initial_retry_backoff` is used for calculating backoff and the backoff is not higher than `initial_retry_backoff * 2`.

> The value must be passed in the duration format: `<number>(ms|s|m|h)`.

> Deprecated. Use `seq_db.max_retry_backoff` instead.

**`grpc_keepalive_params`** *`GRPCKeepaliveParams`* *`optional`*

If gRPC keepalive params are not set, no keepalive params are applied to gRPC client.

> Deprecated. Use `seq_db.grpc_keepalive_params` instead.

`GRPCKeepaliveParams` fields:

+ **`time`** *`string`* *`default="10s"`*

  After a duration of this time if the client doesn't see any activity it pings the server to see if the transport is still alive. If set below `10s`, then it will be reset to `default`.

  > The value must be passed in the duration format: `<number>(ms|s|m|h)`.

+ **`timeout`** *`string`* *`default="1s"`*

  After having pinged for keepalive check, the client waits for a duration of Timeout and if no activity is seen even after that the connection is closed. If set below `1s`, then it will be reset to `default`.

  > The value must be passed in the duration format: `<number>(ms|s|m|h)`.

+ **`permit_without_stream`** *`bool`* *`default=false`*

  If set to `true`, client sends keepalive pings even with no active RPCs. Otherwise, when there are no active RPCs, `time` and `timeout` will be ignored and no keepalive pings will be sent.

**`seq_db`** *`[]SeqDBClient`* *`optional`*

List of SeqDB client configurations.

`SeqDBClient` fields:

+ **`id`** *`string`* *`required`*

  Unique client identifier.

+ **`addrs`** *`[]string`* *`required`*

  List of seq-db proxy hosts to be used in client calls. If there are more than one host, for each request random host will be chosen.

+ **`timeout`** *`string`* *`default="0"`*

  Timeout for requests made by the client. A zero value means no timeout.

  > The value must be passed in the duration format: `<number>(ms|s|m|h)`.

+ **`avg_doc_size`** *`int`* *`default=0`*

  Specifies the average documents size in `KB` that the client calls returns. It's used in combination with `handlers.seq_api.max_search_limit` to calculate the maximum response size per client request.

  > Regardless of `avg_doc_size`, the minimum response size per client request is `4MB`.

+ **`request_retries`** *`int`* *`default=0`*

  The number of retries to send a request to client after the first attempt. For each retry, the entire `addrs` list will be searched, choosing a random host. A zero value means no retries. If set to negative value, then it will be reset to `default`.

+ **`initial_retry_backoff`** *`string`* *`default="0"`*

  Initial duration interval value to be used in backoff with retries. If set to `0`, disables backoff.

  > The value must be passed in the duration format: `<number>(ms|s|m|h)`.

+ **`max_retry_backoff`** *`string`* *`default="0"`*

  Max duration interval value to be used in backoff with retries. If set to `0`, only value from `initial_retry_backoff` is used for calculating backoff and the backoff is not higher than `initial_retry_backoff * 2`.

  > The value must be passed in the duration format: `<number>(ms|s|m|h)`.

+ **`client_mode`** *`string`* *`default="grpc"`* *`options="grpc"`

  Allow choosing how to send requests to seq-db.

+ **`grpc_keepalive_params`** *`GRPCKeepaliveParams`* *`optional`*

  If gRPC keepalive params are not set, no keepalive params are applied to gRPC client.

## Handlers

```yaml
handlers:
  seq_api:
  error_groups:
  mass_export:
```

### SeqAPI

**`seq_api`** *`SeqAPI`* *`optional`*

Config for `/seqapi` API handlers.

`SeqAPI` fields:

+ **`max_search_limit`** *`int`* *`default=0`*

  Max value for `limit` field in search requests.

+ **`max_search_total_limit`** *`int`* *`default=1e6`*

  If search request returns number of events greater than `max_search_total_limit`, then the error will return.

+ **`max_search_offset_limit`** *`int`* *`default=1e6`*

  Max value for `offset` field in search requests.

+ **`max_export_limit`** *`int`* *`default=0`*

  Max value for `limit` field in export requests.

+ **`seq_cli_max_search_limit`** *`int`* *`default=0`*

  The maximum number of logs that can be processed by seq-cli in one search command run.

+ **`max_parallel_export_requests`** *`int`* *`default=1`*

  Number of parallel export requests allowed. If [auth](#auth) is active, each user gets personal quota, otherwise the quota is general for all requests. If set to zero or negative value, then it will be reset to `default`.

+ **`max_aggregations_per_request`** *`int`* *`default=1`*

  Max allowed aggregations per request. If set to zero or negative value, then it will be reset to `default`.

+ **`max_buckets_per_aggregation_ts`** *`int`* *`default=200`*

  Max allowed buckets per aggregation with timeseries request. The number of buckets is calculated as (`to`-`from`) / `interval`. If set to zero or negative value, then it will be reset to `default`.

+ **`events_cache_ttl`** *`string`* *`default="24h"`*

  TTL for events caching. If not set or set to zero, then it will be reset to `default`.

  > The value must be passed in the duration format: `<number>(ms|s|m|h)`.

+ **`logs_lifespan_cache_key`** *`string`* *`default="logs_lifespan"`*

  Cache key for logs lifespan data. Useful when multiple instances of seq-ui use one Redis cache. If set to empty string, then it will be reset to `default`.

+ **`logs_lifespan_cache_ttl`** *`string`* *`default="10m"`*

  TTL for logs lifespan caching. If not set or set to zero, then it will be reset to `default`.

  > The value must be passed in the duration format: `<number>(ms|s|m|h)`.

+ **`fields_cache_ttl`** *`string`* *`default="0"`*

  TTL for fields caching. A zero value means no caching.

  > The value must be passed in the duration format: `<number>(ms|s|m|h)`.

+ **`pinned_fields`** *`[]PinnedField`* *`default=[]`*

  List of fields which will be pinned in UI.

  `PinnedField` fields:

  + **`name`** *`string`* *`required`*

    Name of field.

  + **`type`** *`string`* *`required`* *`options="text"|"keyword"`*

    Type of field.

+ **`masking`** *`Masking`* *`optional`*

  Masking configuration.

  ⚠ **Experimental feature**

  `Masking` fields:

  + **`masks`** *`[]Mask`* *`required`*

    List of masks.

  + **`process_fields`** *`[]string`* *`default=[]`*

    List of processed event fields.

    > It is wrong to set non-empty ignored fields list and non-empty processed fields list at the same time.
  
  + **`ignore_fields`** *`[]string`* *`default=[]`*

    List of ignored event fields.

    > It is wrong to set non-empty ignored fields list and non-empty processed fields list at the same time.
  
  `Mask` fields:

  + **`re`** *`string`* *`required`*

    Regular expression for masking.

  + **`groups`** *`[]int`* *`default=[]`*

    Groups are numbers of masking groups in expression. If set to empty list or the list **contains** `0`, the full expression will be masked.
  
  + **`mode`** *`string`* *`required`* *`options="mask"|"replace"|"cut"`*

    Masking mode:
      - `mask` - asterisks (`*`) are used for masking
      - `replace` - `replace_word` is used for masking
      - `cut` - masking parts will be cut instead of being replaced

  + **`replace_word`** *`string`* *`default=""`*

    Replacement word used in `mode: replace`.

    > Ignored in other mods.
  
  + **`process_fields`** *`[]string`* *`default=[]`*

    List of mask-specific processed event fields.

    > It is wrong to set non-empty ignored fields list and non-empty processed fields list at the same time.
  
  + **`ignore_fields`** *`[]string`* *`default=[]`*

    List of mask-specific ignored event fields.

    > It is wrong to set non-empty ignored fields list and non-empty processed fields list at the same time.
  
  + **`field_filters`** *`FieldFilterSet`* *`optional`*

    Set of field filters to filter events before masking.

  `FieldFilterSet` fields:

  + **`condition`** *`string`* *`required`* *`options="and"|"or"|"not"`*

    Condition for combining filters.

  + **`filters`** *`[]FieldFilter`* *`required`*

    List of filters.

    > Maximum 1 when `condition: not`.
  
  `FieldFilter` fields:

  + **`field`** *`string`* *`required`*

    Event field.

  + **`mode`** *`string`* *`required`* *`options="equal"|"contains"|"prefix"|"suffix"`*

    Filter mode.

  + **`values`** *`[]string`* *`required`*

    List of event field values to filter.
  
  
+ **`envs`** *`map[string]SeqAPIEnv`* *`optional`*

  Environment-specific configurations

  `SeqAPIEnv` fields:

  + **`seq_db_id`** *`string`* *`required`*
  
    SeqDB client identifier from `clients.seq_db` configuration.

  + **`options`** *`SeqAPIOptions`* *`optional`*
  
  Environment-specific API parameters. These settings override the corresponding global values defined in the root `seq_api` section.

### Error groups

**`error_groups`** *`ErrorGroups`* *`optional`*

Config for `/errorgroups` API handlers.

`ErrorGroups` fields:

+ **`log_tags_mapping`** *`LogTagsMapping`* *`optional`*

  Mapping of clickhouse column names and `log_tags` keys.

  `LogTagsMapping` fields:

  + **`release`** *`[]string`* *`default=[]`*

    `log_tags` keys for `release` column.
  
  + **`env`** *`[]string`* *`default=[]`*

    `log_tags` keys for `env` column.

+ **`query_filter`** *`map[string]string`* *`optional`*

  Additional conditions to be added to clickhouse queries.

### Mass export

**`mass_export`** *`MassExport`* *`optional`*

Config for `/massexport` API handlers.

`MassExport` fields:

+ **`batch_size`** *`int`* *`default=10000`*

  Size of batch to fetch logs from log storage per request.

+ **`workers_count`** *`int`* *`required`*

  Number of workers downloading logs from `seq-db` and uploading them to file store simultaneously. Must be positive.

+ **`tasks_channel_size`** *`int`* *`default=10000000`*

  Size of channel that contains time subsegments to export. Must be positive.

+ **`part_length`** *`string`* *`default="1h"`*

  Length of time segment which logs stored in one file.

  > The value must be passed in the duration format: `<number>(ms|s|m|h)`.

+ **`url_prefix`** *`string`* *`required`*

  URL prefix to form links to files in s3. Must be non-empty string.

+ **`allowed_users`** *`[]string`* *`default=[]`*

  List of users who can use `/massexport` API.. If it's empty then mass exports allowed for all users; display username is `anonymous`.

+ **`file_store`** *`FileStore`* *`required`*

  File store config.

+ **`session_store`** *`SessionStore`* *`required`*

  Session store config.

+ **`seq_proxy_downloader`** *`SeqProxyDownloader`* *`required`*

  seq-db proxy client config.

`FileStore` fields:

+ **`s3`** *`S3`* *`required`*

  S3 config.

  `S3` fields:
  + **`endpoint`** *`string`* *`required`*
  + **`access_key_id`** *`string`* *`required`*
  + **`secret_access_key`** *`string`* *`required`*
  + **`bucket_name`** *`string`* *`required`*
  + **`enable_ssl`** *`bool`* *`default=false`*

`SessionStore` fields:

+ **`redis`** *`Redis`* *`required`*

  Redis session store config. See `Redis` in [Cache](#cache) section.

+ **`export_lifetime`** *`string`* *`default="168h"`*

  Expiration time for all keys stored in redis.

`SeqProxyDownloader` fields:

+ **`delay`** *`string`* *`default="1s"`*

  Represents delay between seq-db search queries.

  > The value must be passed in the duration format: `<number>(ms|s|m|h)`.

+ **`initial_retry_backoff`** *`string`* *`default="0"`*

  Initial retry backoff if previous query was rate-limited. If it's less than `delay`, then `delay` value will be reset to `initial_retry_backoff`.
  
  > The value must be passed in the duration format: `<number>(ms|s|m|h)`.

+ **`max_retry_backoff`** *`string`* *`default="0"`*

  Max retry backoff if previous query was rate-limited. If it's less than `initial_retry_backoff`, then `initial_retry_backoff` value will be reset to `max_retry_backoff`

  > The value must be passed in the duration format: `<number>(ms|s|m|h)`.

## Tracing

The tracing configuration is set through environment variables.

```bash
export TRACING_AGENT_HOST=localhost
export TRACING_AGENT_PORT=6831
export TRACING_SAMPLER_PARAM=0.1
export TRACING_SERVICE_NAME=seq-ui
```

### Field Details

+ **`TRACING_SERVICE_NAME`**  *`string`* *`required`*

  Identifies the service name in tracing systems.

+ **`TRACING_AGENT_HOST`** *`string`* *`required`*
  
  Defines the host address of the tracing agent (e.g., Jaeger). 

+ **`TRACING_AGENT_PORT`** *`string`* *`required`*
  
  Port of the tracing agent.

+ **`TRACING_SAMPLER_PARAM`** *`float64`* *`required`*
  
  Sampling rate parameter. Determines the fraction of requests that will be traced. Must be value between 0.0 and 1.0. For instance, use 0.1 to sample 10% of requests.
