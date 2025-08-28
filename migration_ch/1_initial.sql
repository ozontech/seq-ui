-- +goose Up
CREATE TABLE IF NOT EXISTS seq_ui_server.events_raw
(
    timestamp DateTime,
    service String,
    _group_hash UInt64,
    env LowCardinality(String),
    source LowCardinality(String),
    cluster LowCardinality(String),
    release String,
    message String,
    log_tags Map(String, String),
    ttl DateTime DEFAULT now()
)
ENGINE MergeTree()
PARTITION BY toStartOfTenMinutes(timestamp)
ORDER BY (timestamp)
TTL ttl + INTERVAL 10 MINUTES
SETTINGS ttl_only_drop_parts = 1, merge_with_ttl_timeout = 1800, index_granularity = 8192;

CREATE TABLE IF NOT EXISTS seq_ui_server.error_groups
(
    _group_hash UInt64,
    service String,
    env LowCardinality(String),
    source LowCardinality(String),
    cluster LowCardinality(String),
    release String,
    message String,
    seen_total AggregateFunction(count),
    first_seen_at AggregateFunction(min, DateTime),
    last_seen_at AggregateFunction(max, DateTime),
    log_tags Map(String, String),
    ttl DateTime
)
ENGINE = AggregatingMergeTree()
ORDER BY (cluster, source, env, service, release, _group_hash);

CREATE MATERIALIZED VIEW IF NOT EXISTS seq_ui_server.error_groups_mv TO seq_ui_server.error_groups
AS SELECT
    _group_hash,
    service,
    env,
    source,
    cluster,
    release,
    any(message) as message,
    minState(toDateTime(timestamp)) as first_seen_at,
    maxState(toDateTime(timestamp)) as last_seen_at,
    countState() as seen_total,
    any(log_tags) as log_tags,
    max(timestamp) as ttl
FROM seq_ui_server.events_raw
GROUP BY _group_hash, service, env, release, source, cluster;

CREATE TABLE IF NOT EXISTS seq_ui_server.agg_events_10min (
    start_date DateTime NOT NULL,
    service String,
    _group_hash UInt64,
    env LowCardinality(String),
    source LowCardinality(String),
    cluster LowCardinality(String),
    release String,
    counts AggregateFunction(count)
)
ENGINE = AggregatingMergeTree()
PARTITION BY toStartOfHour(start_date)
ORDER BY (cluster, source, env, service, release, _group_hash, start_date)
TTL start_date + INTERVAL 3 MONTH
SETTINGS ttl_only_drop_parts = 1, merge_with_ttl_timeout = 1800, index_granularity = 8192;

CREATE MATERIALIZED VIEW IF NOT EXISTS seq_ui_server.agg_events_10min_mv TO seq_ui_server.agg_events_10min
AS SELECT
    toStartOfTenMinutes(timestamp) as start_date,
    service,
    _group_hash,
    env,
    source,
    cluster,
    release,
    countState() AS counts
FROM seq_ui_server.events_raw
GROUP BY start_date, _group_hash, service, env, release, source, cluster;

CREATE TABLE IF NOT EXISTS seq_ui_server.services (
    env LowCardinality(String),
    cluster LowCardinality(String),
    service String,
    release String,
    ttl DateTime
) ENGINE = ReplacingMergeTree()
ORDER BY (cluster, env, service, release);

CREATE MATERIALIZED VIEW IF NOT EXISTS seq_ui_server.services_mv TO seq_ui_server.services
AS SELECT
    env,
    cluster,
    service,
    release,
    max(timestamp) as ttl
FROM seq_ui_server.events_raw
GROUP BY cluster, env, service, release;

-- +goose Down
DROP TABLE IF EXISTS seq_ui_server.events_raw;
DROP TABLE IF EXISTS seq_ui_server.error_groups;
DROP TABLE IF EXISTS seq_ui_server.error_groups_mv;
DROP TABLE IF EXISTS seq_ui_server.agg_events_10min;
DROP TABLE IF EXISTS seq_ui_server.services;
DROP TABLE IF EXISTS seq_ui_server.services_mv;
