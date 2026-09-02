-- +goose Up
ALTER TABLE seq_ui_server.events_raw
    ADD COLUMN IF NOT EXISTS attributes Map(String, String) AFTER log_tags;

ALTER TABLE seq_ui_server.error_groups
    ADD COLUMN IF NOT EXISTS attributes SimpleAggregateFunction(anyLast, Array(String)),
    ADD COLUMN IF NOT EXISTS attributes_hash UInt64,
    ADD INDEX IF NOT EXISTS idx_attributes attributes TYPE bloom_filter(0.01) GRANULARITY 1,
    MODIFY ORDER BY (cluster, source, env, service, release, _group_hash, attributes_hash);

ALTER TABLE seq_ui_server.error_groups_mv MODIFY QUERY
WITH
    arrayMap(
        (k, v) -> concat(k, '=', v),
        mapKeys(attributes),
        mapValues(attributes)
    ) AS attrs_array
SELECT
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
    attrs_array as attributes,
    xxHash64(arraySort(attrs_array)) AS attributes_hash
FROM seq_ui_server.events_raw
GROUP BY _group_hash, service, env, source, cluster, release, attributes;

ALTER TABLE seq_ui_server.agg_events_10min
    ADD COLUMN IF NOT EXISTS attributes SimpleAggregateFunction(anyLast, Array(String)),
    ADD COLUMN IF NOT EXISTS attributes_hash UInt64,
    ADD INDEX IF NOT EXISTS idx_attributes attributes TYPE bloom_filter(0.01) GRANULARITY 1,
    MODIFY ORDER BY (cluster, source, env, service, release, _group_hash, start_date, attributes_hash);

ALTER TABLE seq_ui_server.agg_events_10min_mv MODIFY QUERY
WITH
    arrayMap(
        (k, v) -> concat(k, '=', v),
        mapKeys(attributes),
        mapValues(attributes)
    ) AS attrs_array
SELECT
    toStartOfTenMinutes(timestamp) as start_date,
    service,
    _group_hash,
    env,
    source,
    cluster,
    release,
    countState() AS counts,
    attrs_array as attributes,
    xxHash64(arraySort(attrs_array)) AS attributes_hash
FROM seq_ui_server.events_raw
GROUP BY start_date, _group_hash, service, env, release, source, cluster, attributes;

ALTER TABLE seq_ui_server.agg_events_1d
    ADD COLUMN IF NOT EXISTS attributes SimpleAggregateFunction(anyLast, Array(String)),
    ADD COLUMN IF NOT EXISTS attributes_hash UInt64,
    ADD INDEX IF NOT EXISTS idx_attributes attributes TYPE bloom_filter(0.01) GRANULARITY 1,
    MODIFY ORDER BY (cluster, source, env, service, release, _group_hash, start_date, attributes_hash);

ALTER TABLE seq_ui_server.agg_events_1d_mv MODIFY QUERY
SELECT
    toStartOfDay(start_date) as start_date,
    service,
    _group_hash,
    env,
    source,
    cluster,
    release,
    countMergeState(counts) AS counts,
    anyLast(attributes) as attributes,
    attributes_hash
FROM seq_ui_server.agg_events_10min
GROUP BY start_date, _group_hash, service, env, release, source, cluster, attributes_hash;

CREATE TABLE IF NOT EXISTS seq_ui_server.service_attributes
(
    cluster LowCardinality(String),
    env LowCardinality(String),
    service String,
    key String,
    values AggregateFunction(groupUniqArray, String)
)
ENGINE = AggregatingMergeTree()
ORDER BY (cluster, env, service, key);

CREATE MATERIALIZED VIEW IF NOT EXISTS seq_ui_server.service_attributes_mv TO seq_ui_server.service_attributes
AS SELECT
    cluster,
    env,
    service,
    key,
    groupUniqArrayState(value) AS values
FROM seq_ui_server.events_raw
ARRAY JOIN mapKeys(attributes) AS key, mapValues(attributes) AS value
GROUP BY cluster, env, service, key;

ALTER TABLE seq_ui_server.error_groups
    REMOVE TTL,
    DROP COLUMN IF EXISTS ttl;

CREATE TABLE IF NOT EXISTS seq_ui_server.error_groups_ttl
(
    _group_hash UInt64,
    cluster LowCardinality(String),
    source LowCardinality(String),
    env LowCardinality(String),
    service String,
    ttl DateTime
)
ENGINE = ReplacingMergeTree(ttl)
ORDER BY (cluster, source, env, service, _group_hash)
TTL ttl + INTERVAL 3 MONTH;

CREATE MATERIALIZED VIEW IF NOT EXISTS seq_ui_server.error_groups_ttl_mv TO seq_ui_server.error_groups_ttl
AS SELECT
    _group_hash,
    cluster,
    source,
    env,
    service,
    max(timestamp) as ttl
FROM seq_ui_server.events_raw
GROUP BY _group_hash, cluster, source, env, service;

ALTER TABLE seq_ui_server.error_groups_brief_mv MODIFY QUERY
SELECT
    _group_hash,
    cluster,
    source,
    env,
    countState() as seen_total
FROM seq_ui_server.events_raw
GROUP BY cluster, source, env, _group_hash;

ALTER TABLE seq_ui_server.error_groups_brief
    REMOVE TTL,
    DROP COLUMN IF EXISTS ttl;

CREATE TABLE IF NOT EXISTS seq_ui_server.error_groups_brief_ttl
(
    _group_hash UInt64,
    cluster LowCardinality(String),
    source LowCardinality(String),
    env LowCardinality(String),
    ttl DateTime
)
ENGINE = ReplacingMergeTree(ttl)
ORDER BY (cluster, source, env, _group_hash)
TTL ttl + INTERVAL 3 MONTH;

CREATE MATERIALIZED VIEW IF NOT EXISTS seq_ui_server.error_groups_brief_ttl_mv TO seq_ui_server.error_groups_brief_ttl
AS SELECT
    _group_hash,
    cluster,
    source,
    env,
    max(timestamp) as ttl
FROM seq_ui_server.events_raw
GROUP BY _group_hash, cluster, source, env;

ALTER TABLE seq_ui_server.services MODIFY TTL ttl + INTERVAL 2 YEAR;

-- +goose Down
ALTER TABLE seq_ui_server.services MODIFY TTL ttl + INTERVAL 3 MONTH;

DROP TABLE IF EXISTS seq_ui_server.error_groups_brief_ttl_mv;
DROP TABLE IF EXISTS seq_ui_server.error_groups_brief_ttl;

ALTER TABLE seq_ui_server.error_groups_brief
    ADD COLUMN IF NOT EXISTS ttl DateTime,
    MODIFY TTL ttl + INTERVAL 3 MONTH;

ALTER TABLE seq_ui_server.error_groups_brief_mv MODIFY QUERY
SELECT
    _group_hash,
    cluster,
    source,
    env,
    countState() as seen_total,
    max(timestamp) as ttl
FROM seq_ui_server.events_raw
GROUP BY cluster, source, env, _group_hash;

DROP TABLE IF EXISTS seq_ui_server.error_groups_ttl_mv;
DROP TABLE IF EXISTS seq_ui_server.error_groups_ttl;

ALTER TABLE seq_ui_server.error_groups
    ADD COLUMN IF NOT EXISTS ttl DateTime,
    MODIFY TTL ttl + INTERVAL 3 MONTH;

DROP TABLE IF EXISTS seq_ui_server.service_attributes_mv;
DROP TABLE IF EXISTS seq_ui_server.service_attributes;

ALTER TABLE seq_ui_server.agg_events_1d_mv MODIFY QUERY
SELECT
    toStartOfDay(start_date) as start_date,
    service,
    _group_hash,
    env,
    source,
    cluster,
    release,
    countMergeState(counts) AS counts
FROM seq_ui_server.agg_events_10min
GROUP BY start_date, _group_hash, service, env, release, source, cluster;

-- we can't roll back the ORDER BY change
ALTER TABLE seq_ui_server.agg_events_1d
    DROP INDEX IF EXISTS idx_attributes,
    DROP COLUMN IF EXISTS attributes;

ALTER TABLE seq_ui_server.agg_events_10min_mv MODIFY QUERY
SELECT
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

-- we can't roll back the ORDER BY change
ALTER TABLE seq_ui_server.agg_events_10min
    DROP INDEX IF EXISTS idx_attributes,
    DROP COLUMN IF EXISTS attributes;

ALTER TABLE seq_ui_server.error_groups_mv MODIFY QUERY
SELECT
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
GROUP BY _group_hash, service, env, source, cluster, release;

-- we can't roll back the ORDER BY change
ALTER TABLE seq_ui_server.error_groups
	DROP INDEX IF EXISTS idx_attributes,
	DROP COLUMN IF EXISTS attributes;

ALTER TABLE seq_ui_server.events_raw
    DROP COLUMN IF EXISTS attributes;
