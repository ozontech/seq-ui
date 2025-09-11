# Seq API
Seq API provides:
- access to logs that are stored in [seq-db](https://github.com/ozontech/seq-db/)
- calculation aggregations and histograms based on logs
- access to seq-ui limits and seq-db stores status

## HTTP API

**Base URL:** `/seqapi/v1`

> You can also use [swagger file](https://github.com/ozontech/seq-ui/blob/main/swagger/swagger.json) to view the HTTP API in detail.

### `GET /fields`

Returns a list of [indexed fields](https://github.com/ozontech/seq-db/blob/main/docs/en/03-index-types.md), specified in the seq-db mapping file.

**Auth:** NO

#### Request

```shell
curl -X GET \
  "http://localhost:5555/seqapi/v1/fields" \
  -H "accept: application/json"
```

#### Response

```json
{
  "fields": [
    {
      "name": "message",
      "type": "text"
    },
    {
      "name": "level",
      "type": "keyword"
    }
  ]
}
```

### `POST /search`

Returns a list of events that satisfy the [search query](https://github.com/ozontech/seq-db/blob/main/docs/en/05-seq-ql.md).

> Allows obtain [aggregations](#aggregation) and a [histogram](#histogram) within a single query.

**Auth:** YES

**Request Body (application/json):**
- `query` (*string*, *optional*): Search query.
- `from` (*string*, *required*): Timestamp of the start of search in `date-time` format.
- `to` (*string*, *required*): Timestamp of the end of search in `date-time` format.
- `histogram` (*object*, *optional*): Histogram query.
  - `interval` (*string*, *required*): Histogram interval in `duration` format.
- `aggregations` (*[]object*, *optional*): List of aggregation queries (see [/aggregation](#aggregation) for details).
- `limit` (*int*, *required*): Search limit.
- `offset` (*int*, *optional*): Search offset.
- `withTotal` (*bool*, *optional*): If set, returns the total number of events found.
- `order` (*enum*, *optional*): Search order. One of `"desc"|"asc"` (`"desc"` by default).

#### Request

```shell
curl -X POST \
  "http://localhost:5555/seqapi/v1/search" \
  -H "accept: application/json" \
  -H "Content-Type: application/json" \
  -d '
  {
    "query": "message:error or level:3",
    "from": "2025-09-10T07:00:00Z",
    "to": "2025-09-10T08:00:00Z",
    "limit": 3,
    "offset": 0,
    "withTotal": true
  }'
```

#### Response
```json
{
  "events": [
    {
      "id": "a78ea33299010000-4101e1b86db21cc9",
      "data": {
        "level": "6",
        "message": "no error",
        "timestamp": "2025-09-10 07:54:03.862"
      },
      "time": "2025-09-10T07:54:03.862Z"
    },
    {
      "id": "a78ea33299010000-410190f323c58b98",
      "data": {
        "level": "3",
        "message": "Unexpected packet Data received from client",
        "timestamp": "2025-09-10 07:50:02.123"
      },
      "time": "2025-09-10T07:50:02.123Z"
    },
    {
      "id": "a78ea33299010000-4101ee1666b9be8f",
      "data": {
        "level": "3",
        "message": "Too many parts",
        "timestamp": "2025-09-10 07:42:12.456"
      },
      "time": "2025-09-10T07:42:12.456Z"
    }
  ],
  "total": "52",
  "error": {
    "code": "ERROR_CODE_NO"
  }
}
```

### `GET /events/{id}`

Retrieves a specific event by their ID.

**Auth:** YES

**Params:**
- `id` (*string*, *required*): The unique identifier of the event.

#### Request

```shell
curl -X GET \
  "http://localhost:5555/seqapi/v1/events/a78ea33299010000-410190f323c58b98" \
  -H "accept: application/json"
```

#### Response
```json
{
  "event": {
    "id": "a78ea33299010000-410190f323c58b98",
    "data": {
      "level": "3",
      "message": "Unexpected packet Data received from client",
      "timestamp": "2025-09-10 07:50:02.123"
    },
    "time": "2025-09-10T07:50:02.123Z"
  }
}
```

### `POST /export`

Downloads events to file in the specified format.

**Auth:** YES

**Request Body (application/json):**
- `format` (*enum*, *optional*): Export format. One of `"jsonl"|"csv"` (`"jsonl"` by default).
- `query` (*string*, *optional*): Search query.
- `from` (*string*, *required*): Timestamp of the start of search in `date-time` format.
- `to` (*string*, *required*): Timestamp of the end of search in `date-time` format.
- `limit` (*int*, *required*): Export limit.
- `offset` (*int*, *optional*): Export offset.
- `fields` (*[]string*, *optional*): List of fields to export (only for `format:csv`, must be non-empty in this case).

#### Request

JSONL:
```shell
curl -X POST \
  "http://localhost:5555/seqapi/v1/export" \
  -H "accept: application/json" \
  -H "Content-Type: application/json" \
  -d '
  {
    "format": "jsonl",
    "query": "message:error or level:3",
    "from": "2025-09-10T07:00:00Z",
    "to": "2025-09-10T08:00:00Z",
    "limit": 3,
    "offset": 0
  }'
```

CSV:
```shell
curl -X POST \
  "http://localhost:5555/seqapi/v1/export" \
  -H "accept: application/json" \
  -H "Content-Type: application/json" \
  -d '
  {
    "format": "csv",
    "fields": ["level", "message"],
    "query": "message:error or level:3",
    "from": "2025-09-10T07:00:00Z",
    "to": "2025-09-10T08:00:00Z",
    "limit": 3,
    "offset": 0
  }'
```

#### Response

Data is returned in chunks. The end of response can be determined by the `Content-Length: 0` header.

JSONL:
```json
{"id": "a78ea33299010000-4101e1b86db21cc9","data": {"level": "6","message": "no error","timestamp": "2025-09-10 07:54:03.862"},"time": "2025-09-10T07:54:03.862Z"}
{"id": "a78ea33299010000-410190f323c58b98","data": {"level": "3","message": "Unexpected packet Data received from client","timestamp": "2025-09-10 07:50:02.123"},"time": "2025-09-10T07:50:02.123Z"}
{"id": "a78ea33299010000-4101ee1666b9be8f","data": {"level": "3","message": "Too many parts","timestamp": "2025-09-10 07:42:12.456"},"time": "2025-09-10T07:42:12.456Z"}
```

CSV:
```
level,message
6,no error
3,Unexpected packet Data received from client
3,Too many parts
```

### `POST /aggregation`

Calculates aggregations based on events that satisfy the search query.

> Aggregations can also be obtained using [/search](#search).

**Auth:** YES

**Request Body (application/json):**
- `query` (*string*, *optional*): Search query.
- `from` (*string*, *required*): Timestamp of the start of search in `date-time` format.
- `to` (*string*, *required*): Timestamp of the end of search in `date-time` format.
- `aggregations` (*[]object*, *required*): List of aggregation queries.
  - `agg_func` (*enum*, *optional*): Aggregation function. One of `"count"|"sum"|"min"|"max"|"avg"|"quantile"|"unique"` (`"count"` by default).
  - `field` (*string*, *required*): Aggregation calculation field.
  - `group_by` (*string*, *optional*): Field for grouping the aggregation results.
  - `quantiles` (*[]int*, *optional*): List of quantiles (only for `agg_func:quantile`, must be non-empty in this case).

#### Request

```shell
curl -X POST \
  "http://localhost:5555/seqapi/v1/aggregation" \
  -H "accept: application/json" \
  -H "Content-Type: application/json" \
  -d '
  {
    "query": "_exists_:level",
    "from": "2025-09-10T07:00:00Z",
    "to": "2025-09-10T08:00:00Z",
    "aggregations": [
      {
        "agg_func": "count",
        "field": "level"
      }
    ]
  }'
```

#### Response
```json
{
  "aggregations": [
    {
      "buckets": [
        {
          "key": "6",
          "value": 203
        },
        {
          "key": "4",
          "value": 20
        },
        {
          "key": "3",
          "value": 18
        }
      ]
    }
  ],
  "error": {
    "code": "ERROR_CODE_NO"
  }
}
```

### `POST /aggregation_ts`

Calculates aggregations within different time intervals (also known as timeseries) based on events that satisfy the search query.

**Auth:** YES

**Request Body (application/json):**
- `query` (*string*, *optional*): Search query.
- `from` (*string*, *required*): Timestamp of the start of search in `date-time` format.
- `to` (*string*, *required*): Timestamp of the end of search in `date-time` format.
- `aggregations` (*[]object*, *required*): List of aggregation queries.
  - `agg_func` (*enum*, *optional*): Aggregation function. One of `"count"|"sum"|"min"|"max"|"avg"|"quantile"|"unique"` (`"count"` by default).
  - `interval` (*string*, *required*): Interval for calculating the bucket in `duration` format.
  - `field` (*string*, *required*): Aggregation calculation field.
  - `group_by` (*string*, *optional*): Field for grouping the aggregation results.
  - `quantiles` (*[]int*, *optional*): List of quantiles (only for `agg_func:quantile`, must be non-empty in this case).

#### Request

```shell
curl -X POST \
  "http://localhost:5555/seqapi/v1/aggregation_ts" \
  -H "accept: application/json" \
  -H "Content-Type: application/json" \
  -d '
  {
    "query": "_exists_:level",
    "from": "2025-09-10T07:00:00Z",
    "to": "2025-09-10T08:00:00Z",
    "aggregations": [
      {
        "agg_func": "count",
        "field": "level",
        "interval": "30m"
      }
    ]
  }'
```

#### Response
```json
{
  "aggregations": [
    {
      "data": {
        "result": [
          {
            "metric": {
              "level": "6"
            },
            "values": [
              {
                "timestamp": 1757487600,
                "value": 100
              },
              {
                "timestamp": 1757489400,
                "value": 200
              }
            ]
          },
          {
            "metric": {
              "level": "3"
            },
            "values": [
              {
                "timestamp": 1757487600,
                "value": 300
              },
              {
                "timestamp": 1757489400,
                "value": 350
              }
            ]
          },
          {
            "metric": {
              "level": "4"
            },
            "values": [
              {
                "timestamp": 1757487600,
                "value": 400
              },
              {
                "timestamp": 1757489400,
                "value": 1000
              }
            ]
          },
        ]
      }
    }
  ]
}
```

### `POST /histogram`

Calculates histogram based on events that satisfy the search query.

> Histogram can also be obtained using [/search](#search).

**Auth:** YES

**Request Body (application/json):**
- `query` (*string*, *optional*): Search query.
- `from` (*string*, *required*): Timestamp of the start of search in `date-time` format.
- `to` (*string*, *required*): Timestamp of the end of search in `date-time` format.
- `interval` (*string*, *required*): Interval for calculating the bucket in `duration` format.

#### Request

```shell
curl -X POST \
  "http://localhost:5555/seqapi/v1/histogram" \
  -H "accept: application/json" \
  -H "Content-Type: application/json" \
  -d '
  {
    "query": "_exists_:level",
    "from": "2025-09-10T07:00:00Z",
    "to": "2025-09-10T08:00:00Z",
    "interval": "15m"
  }'
```

#### Response
```json
{
  "histogram": {
    "buckets": [
      {
        "key": "1757487600000",
        "docCount": "1010"
      },
      {
        "key": "1757488500000",
        "docCount": "1050"
      },
      {
        "key": "1757489400000",
        "docCount": "1030"
      },
      {
        "key": "1757490300000",
        "docCount": "1200"
      },
      {
        "key": "1757491200000",
        "docCount": "21"
      }
    ]
  },
  "error": {
    "code": "ERROR_CODE_NO"
  }
}
```

### `GET /fields/pinned`

Returns the list of fields that will be pinned in UI. Set in the `handlers.seq_api.pinned_fields` [config section](./02-configuration.md#seqapi).

**Auth:** NO

#### Request

```shell
curl -X GET \
  "http://localhost:5555/seqapi/v1/fields/pinned" \
  -H "accept: application/json"
```

#### Response

```json
{
  "fields": [
    {
      "name": "field1",
      "type": "keyword"
    },
    {
      "name": "field2",
      "type": "text"
    }
  ]
}
```

### `GET /limits`

Returns the list of limits set in the `handlers.seq_api` [config section](./02-configuration.md#seqapi).

**Auth:** NO

#### Request

```shell
curl -X GET \
  "http://localhost:5555/seqapi/v1/limits" \
  -H "accept: application/json"
```

#### Response

```json
{
  "maxSearchLimit": 100,
  "maxExportLimit": 10000,
  "maxParallelExportRequests": 1,
  "maxAggregationsPerRequest": 3,
  "seqCliMaxSearchLimit": 10000
}
```

### `GET /logs_lifespan`

Returns the lifespan of logs (in seconds) in seq-db.

**Auth:** NO

#### Request

```shell
curl -X GET \
  "http://localhost:5555/seqapi/v1/logs_lifespan" \
  -H "accept: application/json"
```

#### Response

```json
{
  "lifespan": 4923192
}
```

### `GET /status`

Returns the status of seq-db stores.

**Auth:** NO

#### Request

```shell
curl -X GET \
  "http://localhost:5555/seqapi/v1/status" \
  -H "accept: application/json"
```

#### Response

```json
{
  "oldest_storage_time": "2025-07-15T13:16:00Z",
  "number_of_stores": 2,
  "stores": [
    {
      "host": "seqdb-1:9002",
      "values": {
        "oldest_time": "2025-07-15T15:06:00Z"
      }
    },
    {
      "host": "seqdb-2:9002",
      "values": {
        "oldest_time": "2025-07-15T13:16:00Z"
      }
    }
  ]
}
```