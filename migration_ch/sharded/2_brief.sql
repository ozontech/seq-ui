-- +goose Up
CREATE TABLE IF NOT EXISTS seq_ui_server_replicated.sharded_error_groups_brief
(
    _group_hash UInt64,
    cluster LowCardinality(String),
    source LowCardinality(String),
    env LowCardinality(String),
    seen_total AggregateFunction(count),
    ttl DateTime
)
ENGINE = ReplicatedAggregatingMergeTree('/clickhouse/tables/{shard}/{table}', '{replica}')
ORDER BY (cluster, source, env, _group_hash)
TTL ttl + INTERVAL 3 MONTH;

CREATE MATERIALIZED VIEW IF NOT EXISTS seq_ui_server_replicated.sharded_error_groups_brief_mv TO seq_ui_server_replicated.sharded_error_groups_brief
AS SELECT
    _group_hash,
    cluster,
    source,
    env,
    countState() as seen_total,
    max(timestamp) as ttl
FROM seq_ui_server_replicated.sharded_events_raw
GROUP BY cluster, source, env, _group_hash;

CREATE TABLE IF NOT EXISTS seq_ui_server_replicated.sharded_agg_events_1d
(
    start_date DateTime NOT NULL,
    service String,
    _group_hash UInt64,
    env LowCardinality(String),
    source LowCardinality(String),
    cluster LowCardinality(String),
    release String,
    counts AggregateFunction(count)
)
ENGINE = ReplicatedAggregatingMergeTree('/clickhouse/tables/{shard}/{table}', '{replica}')
PARTITION BY start_date
ORDER BY (cluster, source, env, service, release, _group_hash, start_date)
TTL start_date + INTERVAL 2 YEAR
SETTINGS ttl_only_drop_parts = 1, merge_with_ttl_timeout = 1800;

CREATE MATERIALIZED VIEW IF NOT EXISTS seq_ui_server_replicated.sharded_agg_events_1d_mv TO seq_ui_server_replicated.sharded_agg_events_1d
AS SELECT
    toStartOfDay(start_date) as start_date,
    service,
    _group_hash,
    env,
    source,
    cluster,
    release,
    countMergeState(counts) AS counts
FROM seq_ui_server_replicated.sharded_agg_events_10min
GROUP BY cluster, source, env, service, release, _group_hash, start_date;

CREATE TABLE IF NOT EXISTS seq_ui_server_replicated.error_groups_brief AS seq_ui_server_replicated.sharded_error_groups_brief ENGINE = Distributed("seq-ui-server-replicated", seq_ui_server_replicated, sharded_error_groups_brief);
CREATE TABLE IF NOT EXISTS seq_ui_server_replicated.agg_events_1d AS seq_ui_server_replicated.sharded_agg_events_1d ENGINE = Distributed("seq-ui-server-replicated", seq_ui_server_replicated, sharded_agg_events_1d);

ALTER TABLE seq_ui_server_replicated.sharded_error_groups MODIFY TTL ttl + INTERVAL 3 MONTH;
ALTER TABLE seq_ui_server_replicated.sharded_services MODIFY TTL ttl + INTERVAL 3 MONTH;

-- +goose Down
DROP TABLE IF EXISTS seq_ui_server_replicated.sharded_error_groups_brief_mv;
DROP TABLE IF EXISTS seq_ui_server_replicated.sharded_error_groups_brief;
DROP TABLE IF EXISTS seq_ui_server_replicated.error_groups_brief;

DROP TABLE IF EXISTS seq_ui_server_replicated.sharded_agg_events_1d_mv;
DROP TABLE IF EXISTS seq_ui_server_replicated.sharded_agg_events_1d;
DROP TABLE IF EXISTS seq_ui_server_replicated.agg_events_1d;

ALTER TABLE seq_ui_server_replicated.sharded_error_groups REMOVE TTL;
ALTER TABLE seq_ui_server_replicated.sharded_services REMOVE TTL;
