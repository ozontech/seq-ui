
# Config

Config is read from yaml file and consists of three sections: server, clients, handlers.

You can specify your config file when running seq-ui-server by providing it with flag `--config`

```sh
go run ./cmd/seq-ui --config <path-to-config-file>
```

## Config file example

```yaml
server:
  oidc:
    auth_urls:
      - "https://sso.example.com"
    cache_secret_key: "test"
  jwt_secret_key: "top-secret"
  http_addr: "127.0.0.1:5555"
  grpc_addr: "127.0.0.1:5556"
  debug_addr: "127.0.0.1:5557"
  cors:
    allowed_origins:
      - "*"
    allowed_methods:
      - "HEAD"
      - "GET"
      - "POST"
    allowed_headers:
      - "Content-Type"
    max_age: 300
  grpc_connection_timeout: 45s
  http_read_header_timeout: 15s
  http_read_timeout: 30s
  http_write_timeout: 45s
  db:
    name: "postgres"
    host: "localhost"
    port: 5432
    pass: "postgres"
    user: "postgres"
    request_timeout: 5s
    connection_pool_capacity: 32
    use_prepared_statements: false
  clickhouse:
    addrs:
      - "localhost:9000"
    database: "seq_ui_server"
    username: "default"
    password: ""
    dial_timeout: "5s"
    read_timeout: "15s"
  rate_limiters:
    seqapi:
      default:
        rate_per_sec: 3
        max_burst: 2
        store_max_keys: 65536
        per_handler: false
      spec_users:
        spammer:
          rate_per_sec: 1
          max_burst: 0
          per_handler: false
          store_max_keys: 20
    userprofile:
      default:
        rate_per_sec: 3
        max_burst: 2
        store_max_keys: 65536
        per_handler: true
    UserProfileService:
      default:
        rate_per_sec: 3
        max_burst: 2
        store_max_keys: 65536
        per_handler: true
    dashboards:
      default:
        rate_per_sec: 3
        max_burst: 2
        store_max_keys: 65536
  cache:
    inmemory:
      num_counters: 10000000
      max_cost: 1000000
      buffer_items: 64
    redis:
      addr: localhost:6379
      username: ivan
      password: admin
      timeout: 1s
      max_retries: 5
clients:
  seq_db_timeout: 15s
  seq_db_avg_doc_size: 100
  seq_db_addrs:
    - "seqdb.example.com:9004"
  request_retries: 3
  proxy_client_mode: "grpc"
  grpc_keepalive_params:
    time: 10s
    timeout: 10s
    permit_without_stream: true
handlers:
  seq_api:
    max_search_limit: 100
    max_search_total_limit: 1000000
    max_search_offset_limit: 1000000
    max_export_limit: 10000
    seq_cli_max_search_limit: 10000
    max_parallel_export_requests: 1
    max_aggregations_per_request: 3
    events_cache_ttl: 12h
    pinned_fields:
      - name: field1
        type: keyword
      - name: field2
        type: text
```

## Config params

### Server

Config related to grpc and http servers.

Params:

**`http_addr`** *`string`* *`required`*

Host for HTTP server.

**`grpc_addr`** *`string`* *`required`*

Host for gRPC server.

**`cors`** *`CORS`* *`optional`*

HTTP server CORS config. If not set, no CORS settings will be applied.

`CORS` fields:

+ **`allowed_origins`** *`[]string`* *`default=["*"]`*

    A list of origins a cross-domain request can be executed from. If the special "*" value is present in the list, all origins will be allowed.

+ **`allowed_methods`** *`[]string`* *`default=["HEAD", "GET", "POST"]`*

    A list of methods the client is allowed to use with cross-domain requests.

+ **`allowed_headers`** *`[]string`* *`optional`*

    A list of non simple headers the client is allowed to use with cross-domain requests. If the special "*" value is present in the list, all headers will be allowed. "Origin" header is always appended to the list.

+ **`exposed_headers`** *`[]string`* *`optional`*

    Indicates which headers are safe to expose to the API of a CORS API specification.

+ **`allow_credentials`** *`bool`* *`optional`*

    Indicates whether the request can include user credentials like cookies, HTTP authentication or client side SSL certificates.

+ **`max_age`** *`int`* *`optional`*

    Indicates how long (in seconds) the results of a preflight request can be cached.

+ **`options_passthrough`** *`bool`* *`optional`*

    Instructs preflight to let other potential next handlers to process the OPTIONS method. Turn this on if your application handles OPTIONS.

**`oidc`** *`OIDC`* *`optional`*

Open ID Connnect config. If not set, no OIDC verification will be applied.

`OIDC` fields:

+ **`skip_verify`** *`bool`* *`optional`*

  If set, only the issuer and expiration checked locally without requests to `auth_urls`.

+ **`auth_urls`** *`[]string`* *`required`*

  List of OIDC auth URLs for sending verification requests. For each verification, the entire `auth_urls` list will be searched, choosing a URL.
  
+ **`root_ca`** *`string`* *`optional`*

  Path to file with CA root certificate or the certificate itself. If set, it will be passed to OIDC client tls config.

+ **`ca_cert`** *`string`* *`optional`*

  Path to file with CA certificate or the certificate itself. If set, it will be passed to OIDC client tls config.

+ **`private_key`** *`string`* *`optional`*

  Path to file with private key generated with CA certificate or the private key itself. If set, it will be passed to OIDC client tls config.

+ **`ssl_skip_verify`** *`bool`* *`optional`*

  If set, disables security checks on OIDC client.

+ **`allowed_clients`** *`[]string`*  *`optional`*

  List of allowed clients. If set, only the specified clients will be verified. The `Audience` token field is used.

+ **`cache_secret_key`** *`string`* *`optional`*

  If set to non-empty string, OIDC tokens are cached using `cache_secret_key` until the token expires.

**`jwt_secret_key`** *`string`*  *`optional`*

If set, JWT provider is created for API tokens verification.

> API tokens allow access only for `/seqapi/*` routes in HTTP API and `SeqAPIService` service in gRPC API. For other routes/services requiring auth, OIDC check will be performed, so `jwt_secret_key` should be used in pair with `oidc`.

**`grpc_connection_timeout`** *`string`* *`required`*

Sets the timeout for connection establishment (up to and including HTTP/2 handshaking) for all new connections in gRPC server. The value must be passed in format of duration (`<number>(ms|s|m|h)`).

**`http_read_header_timeout`** *`string`* *`required`*

The amount of time allowed to read request headers. If `http_read_header_timeout` is zero, the value of `http_read_timeout` is used. If both are zero, there is no timeout. The value must be passed in format of duration (`<number>(ms|s|m|h)`).

**`http_read_timeout`** *`string`* *`required`*

The maximum duration for reading the entire request, including the body. A zero or negative value means there will be no timeout. The value must be passed in format of duration (`<number>(ms|s|m|h)`).

**`http_write_timeout`** *`string`* *`required`*

The maximum duration after header is read and before timing out writes of the response. A zero or negative value means there will be no timeout. The value must be passed in format of duration (`<number>(ms|s|m|h)`).

**`db`** *`DB`* *`optional`*

Postgres database config. If not set, app works without `/userprofile` handlers.

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

  Sets a timeout for all database requests. The value must be passed in format of duration (`<number>(ms|s|m|h)`).

+ **`use_prepared_statements`** *`bool`* *`optional`*

  If set to `false`, disables the use of postgres prepared statements. If not set, it will be reset to `true`.

**`clickhouse`** *`ClickHouse`* *`optional`*

ClickHouse database config. If not set, app works without `/errorgroups` handlers.

`ClickHouse` fields:

+ **`database`** *`string`* *`required`*

  Database name.

+ **`username`** *`string`* *`required`*

  Database username.

+ **`password`** *`int`* *`required`*

  Database password.

+ **`dial_timeout`** *`int`* *`required`*

  Sets a timeout for database dial. The value must be passed in format of duration (`<number>(ms|s|m|h)`).

+ **`read_timeout`** *`int`* *`required`*

  Sets a timeout for database read. The value must be passed in format of duration (`<number>(ms|s|m|h)`).

+ **`sharded`** *`bool`* *`optional`*

  Indicates whether the clickhouse is sharded. It is used for some queries that depend on the clickhouse scheme.

**`rate_limiters`** *`map[string]ApiRateLimiters`* *`optional`*

gRPC and HTTP server rate limiters configs. If not set, rate limiters will not be applied.

> Map key is http api base route or grpc api service name:
> - `seqapi`, `userprofile`, `dashboards`, `massexport`, `errorgroups`
> - `SeqAPIService`, `UserProfileService`, `DashboardsService`, `MassExportService`, `ErrorGroupsService`


`ApiRateLimiters` fields:

+ **`default`** *`RateLimiter`* *`required`*

Describes default rate limiter for unspecified and unauthorized users:

+ **`spec_users`** *`map[string]RateLimiter`* *`optional`*

  Describes rate limiters for special users of API. Key is username.

`RateLimiter` fields:

+ **`rate_per_sec`** *`int`* *`required`*

    Number of requests allowed per second. If auth interceptor is active, each user gets quota defined in `rate_limiter`, otherwise the quota is general for all requests.

    > The app utilizes rate limiter based on [Generic Cell Rate Algorithm (GCRA)](https://en.wikipedia.org/wiki/Generic_cell_rate_algorithm). If `max_burst` is set to zero, the next request will be allowed after `(1 sec / rate_per_sec)` time (e.g. `rate_per_sec` is set to 2, then the next request will be allowed after 500ms).

+ **`max_burst`** *`int`* *`required`*

    Number of requests that will be allowed to exceed the rate in a single burst.

    > Total amount of requests allowed per second can be higher than `rate_per_sec`, if `max_burst` is greater than zero.

+ **`store_max_keys`** *`int`* *`optional`*

    Max amount of keys to be stored in rate limiter store. If `store_max_keys` is less or equal to zero, amount of keys is considered as unlimited.


+ **`per_handler`** *`bool`* *`default=false`*

    If flag is `true` then every handler will be limited separately 
    else all api handlers will be limited together.

**`cache`** *`Cache`* *`optional`*

gRPC and HTTP server cache config.

`Cache` fields:

+ **`inmemory`** *`InmemoryCache`* *`optional`*

  Config for in-memory cache.

  > If not set, it will be applied with default values.

  `InmemoryCache` fields:

  + **`num_counters`** *`int`* *`default=10000000`*

    The number of counters (keys) to keep that hold access frequency information. It's generally a good idea to have more counters than the max cache capacity, as this will improve eviction accuracy and subsequent hit ratios.

    > If not set or set to less or equal to 0, then it will be reset to `default`.

  + **`max_cost`** *`int`* *`default=1000000`*

    `max_cost` can be considered as the cache capacity.

    > If not set or set to less or equal to 0, then it will be reset to `default`.

  + **`buffer_items`** *`int`* *`default=64`*

    BufferItems determines the size of Get buffers.

    Unless you have a rare use case, using `64` as the `buffer_items` value results in good performance.

    > If not set or set to less or equal to 0, then it will be reset to `default`.

+ **`redis`** *`Redis`* *`optional`*

  Config for redis cache.

  It works in pair with in-memory cache:
  + when a key-value pair sets in redis, it's also sets in in-memory cache;
  + when in-memory cache contains the key, the request to redis doesn't occur;
  + when in-memory cache doesn't contain the key, but redis contains, the redis result sets in in-memory cache;

  > If redis isn't available, it falls back to in-memory cache.

  `Redis` fields:

  + **`addr`** *`string`* *`required`*

    Address in `host:port` format.

  + **`username`** *`string`* *`optional`*

    Username to authenticate the connection when connecting to a Redis 6.0 instance, or greater, that is using the Redis ACL system.

  + **`password`** *`string`* *`optional`*

    Password specified in the requirepass server configuration option (if connecting to a Redis 5.0 instance, or lower), or the User Password when connecting to a Redis 6.0 instance, or greater, that is using the Redis ACL system.

  + **`timeout`** *`string`* *`optional`*

    Read/write timeout (`3s` by default). The value must be passed in format of duration (`<number>(ms|s|m|h)`).

    > If set to `-1`, disables timeout.

  + **`max_retries`** *`int`* *`optional`*

    Maximum number of retries before giving up (`3` by default).

    > If set to `-1`, disables retries.

  + **`min_retry_backoff`** *`string`* *`optional`*

    Minimum backoff between each retry (`8ms` by default). The value must be passed in format of duration (`<number>(ms|s|m|h)`).

    > If set to `-1`, disables backoff.

  + **`max_retry_backoff`** *`string`* *`optional`*

    Maximum backoff between each retry (`512ms` by default). The value must be passed in format of duration (`<number>(ms|s|m|h)`).

    > If set to `-1`, disables backoff.

### Clients

Config related to seqdb clients.

Params:

**`seq_db_timeout`** *`string`* *`required`*

Specifies a time limit for requests made by the client. A Timeout of zero means no timeout. The value must be passed in format of duration (`<number>(ms|s|m|h)`).

**`seq_db_avg_doc_size`** *`int`* *`optional`*

Specifies the average documents size in KB that the client calls returns.
It's used in combination with `handlers.seq_api.max_search_limit` to calculate the maximum response size per client request.

Regardless of `seq_db_avg_doc_size`, the minimum response size per client request is 4 MB.

**`seq_db_addrs`** *`[]string`* *`required`*

List of seqdb ingestor hosts to be used in client calls. If there are more than one host, for each request random host will be chosen.

**`request_retries`** *`int`* *`optional`*

The number of retries to send a request to client after the first attempt. For each retry, the entire `seq_db_addrs` list will be searched, choosing a random host.

> If `request_retries` value < 0, then it will be reset to 0 (no retries).

**`initial_retry_backoff`** *`string`* *`optional`*

Initial duration interval value to be used in backoff with retries. If not set, no backoff is applied.

> If `initial_retry_backoff` value < 0, then it will be reset to 0 which means no backoff.

**`max_retry_backoff`** *`string`* *`optional`*

Max duration interval value to be used in backoff with retries. If not set, only value from `initial_retry_backoff` is used for calculating backoff and the backoff is not higher than `initial_retry_backoff * 2`.

> If `max_retry_backoff` value < 0, then it will be reset to 0 which means to use only `initial_retry_backoff` for backoff.

**`proxy_client_mode`** *`string`* *`default="grpc"`*

This parameter allows choosing how to send requests to seq-db. Currently, there are one supported mode: `grpc` (default).

**`grpc_keepalive_params`** *`GRPCKeepaliveParams`* *`optional`*

If gRPC keepalive params are not set, no keepalive params are applied to gRPC client.

`GRPCKeepaliveParams` fields:

+ **`time`** *`string`* *`default="10s"`*

    After a duration of this time if the client doesn't see any activity it pings the server to see if the transport is still alive. If set below 10s, a minimum value of 10s will be used instead. The value must be passed in format of duration (`<number>(s|m|h)`).

+ **`timeout`** *`string`* *`default="1s"`*

    After having pinged for keepalive check, the client waits for a duration of Timeout and if no activity is seen even after that the connection is closed. If set below 1s, a minimum value of 1s will be used instead. The value must be passed in format of duration (`<number>(s|m|h)`).

+ **`permit_without_stream`** *`bool`* *`default=false`*

    If true, client sends keepalive pings even with no active RPCs. If false, when there are no active RPCs, Time and Timeout will be ignored and no keepalive pings will be sent. False by default.

### Handlers

Config related to the server handlers.

Params:

**`seq_api`** *`SeqAPI`* *`required`*

Config for `/seqapi` server handlers.

`SeqAPI` fields:

+ **`max_search_limit`** *`int`* *`required`*

    Sets max value for `limit` field in search requests.

+ **`max_export_limit`** *`int`* *`required`*

  Sets max value for `limit` field in export requests.

+ **`seq_cli_max_search_limit`** *`int`* *`required`*

  The maximum number of logs that can be processed by seq-cli in one search command run.

+ **`max_parallel_export_requests`** *`int`* *`default=1`*

  Number of parallel export requests allowed. If auth interceptor is active, each user gets personal quota, otherwise the quota is general for all requests.

  > If not set or set to less or equal to 0, then it will be reset to `default`.

+ **`max_aggregations_per_request`** *`int`* *`default=1`*

  Sets max allowed aggregations per request.
  
  > If not set or set to less or equal to 0, then it will be reset to `default`.

+ **`max_buckets_per_aggregation_ts`** *`int`* *`default=200`*

  Sets max allowed buckets per aggregation with timeseries. The number of buckets is calculated as (`to`-`from`) / `interval`. 
  
  > If not set or set to less or equal to 0, then it will be reset to `default`.

+ **`events_cache_ttl`** *`string`* *`default="24h"`*

  Sets ttl for events caching. The value must be passed in format of duration (`<number>(ms|s|m|h)`).

  > If not set or set to less or equal to 0, then it will be reset to `default`.

+ **`logs_lifespan_cache_key`** *`string`* *`default="logs_lifespan"`*

  Cache key for logs lifespan data. Useful when multiple instances of seq-ui use one Redis cache.

  > If not set or set to empty string, then it will be reset to `default`.

+ **`logs_lifespan_cache_ttl`** *`string`* *`default="10m"`*

  Sets ttl for logs lifespan caching. The value must be passed in format of duration (`<number>(ms|s|m|h)`).

  > If not set or set to less or equal to 0, then it will be reset to `default`.

+ **`fields_cache_ttl`** *`string`*

  Sets ttl for fields caching. The value must be passed in format of duration (`<number>(ms|s|m|h)`).

  > Set positive value to enable caching. By default cache disabled.

+ **`pinned_fields`** *`[]PinnedField`* *`optional`*

  List of fields which will be pinned.

  `PinnedField` fields:

  + **`name`** *`string`* *`required`*

  Name of field.

  + **`type`** *`string`* *`required`*

  Type of field - one of `text`/`keyword` values.

**`error_groups`** *`ErrorGroups`* *`optional`*

Config for `/errorgroups` server handlers.

`ErrorGroups` fields:

+ **`log_tags_mapping`** *`LogTagsMapping`* *`optional`*

    Mapping of clickhouse column names and `log_tags` keys.
  
  `LogTagsMapping` fields:

  + **`release`** *`[]string`* *`required`*

    `log_tags` keys for `release` column.

+ **`query_filter`** *`map[string]string`* *`optional`*

    Additional conditions to be added to clickhouse queries.

**`mass_export`** *`MassExport`* *`optional`*

Config for `/massexport` server handlers.

`MassExport` fields:

+ **`batch_size`** *`int`* *`default=10000`*

  Size of batch to fetch logs from log storage per request.

+ **`workers_count`** *`int`* *`required`*

  Number of workers downloading logs from `seq-db` and uploading them to file store simultaneously. Must be positive.

+ **`tasks_channel_size`** *`int`* *`default=10000000`*

  Size of channel that contains time subsegments to export. Must be positive.

+ **`part_length`** *`string`* *`default=1h`*

  Length of time segment which logs stored in one file. The value must be passed in format of duration (`<number>(ms|s|m|h)`).

+ **`url_prefix`** *`string`* *`required`*

  URL prefix to form links to files in s3. Must be non-empty.

+ **`allowed_users`** *`[]string`* *`required`*

  List of users who can use mass export api. If it's empty then mass exports allowed for all users; display username is `anonymous`.

+ **`file_store`** *`FileStore`* *`required`*

  File store config

+ **`session_store`** *`SessionStore`* *`required`*

  Session store config

+ **`seq_proxy_downloader`** *`SeqProxyDownloader`* *`required`*

  Seq-proxy calls config

`FileStore` fields:

+ **`s3`** *`S3`* *`required`*

  File store implementation config

`S3` fields (credentials):
+ **`endpoint`** *`string`* *`required`*
+ **`access_key_id`** *`string`* *`required`*
+ **`secret_access_key`** *`string`* *`required`*
+ **`bucket_name`** *`string`* *`required`*
+ **`enable_ssl`** *`bool`* *`optional`* *`default=false`*

`SessionStore` fields:

+ **`redis`** *`Redis`* *`required`*

  Redis session store implementation config

+ **`export_lifetime`** *`string`* *`default=168h`*

  Expiration time for all keys stored in redis.

`SeqProxyDownloader` fields:

+ **`delay`** *`string`* *`default=1s`*

  Represents delay between seq-db search queries. The value must be passed in format of duration (`<number>(ms|s|m|h)`).

+ **`initial_retry_backoff`** *`string`*

  Initial retry backoff if previous query was rate-limited. 
  The value must be passed in format of duration (`<number>(ms|s|m|h)`).
  
  > If it's less than `delay` then `delay` value will be reset to `initial_retry_backoff`.

+ **`max_retry_backoff`** *`string`*

  Max retry backoff if previous query was rate-limited. 
  The value must be passed in format of duration (`<number>(ms|s|m|h)`).
  
  > If it's less than `initial_retry_backoff` then `initial_retry_backoff` value will be reset to `max_retry_backoff`.
