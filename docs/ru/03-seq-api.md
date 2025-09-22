# Seq API
Seq API предоставляет:
- доступ к логам, хранящимся в [seq-db](https://github.com/ozontech/seq-db/)
- расчет агрегаций и гистограмм на основе логов
- доступ к ограничениям seq-ui и состоянию хранилищ seq-db

## HTTP API

**Базовый URL-адрес:** `/seqapi/v1`

> Вы также можете использовать [swagger-файл](https://github.com/ozontech/seq-ui/blob/main/swagger/swagger.json) для подробного просмотра HTTP API.

### `GET /fields`

Возвращает список [индексированных полей](https://github.com/ozontech/seq-db/blob/main/docs/ru/03-index-types.md), указанных в mapping-файле seq-db.

**Авторизация:** НЕТ

#### Запрос

```shell
curl -X GET \
  "http://localhost:5555/seqapi/v1/fields" \
  -H "accept: application/json"
```

#### Ответ

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

Возвращает список событий, удовлетворяющих [поисковому запросу](https://github.com/ozontech/seq-db/blob/main/docs/ru/05-seq-ql.md).

> Позволяет получить [агрегации](#post-aggregation) и [гистограмму](#post-histogram) в рамках одного запроса.

**Авторизация:** ДА

**Тело запроса (application/json):**
- `query` (*string*, *optional*): Поисковый запрос.
- `from` (*string*, *required*): Временная метка начала поиска в `date-time` формате.
- `to` (*string*, *required*): Временная метка окончания поиска в `date-time` формате.
- `histogram` (*object*, *optional*): Запрос гистограммы.
  - `interval` (*string*, *required*): Интервал гистограммы в `duration` формате.
- `aggregations` (*[]object*, *optional*): Список запросов на агрегацию (см. [/aggregation](#post-aggregation) для подробностей).
- `limit` (*int*, *required*): Ограничение поиска.
- `offset` (*int*, *optional*): Смещение поиска.
- `withTotal` (*bool*, *optional*): Если задано, то возвращает общее количество найденных событий.
- `order` (*enum*, *optional*): Порядок поиска. Одно из `"desc"|"asc"` (по умолчанию `"desc"`).

#### Запрос

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

#### Ответ

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

Возвращает конкретное событие по его идентификатору.

**Авторизация:** ДА

**Параметры:**
- `id` (*string*, *required*): Уникальный идентификатор события.

#### Запрос

```shell
curl -X GET \
  "http://localhost:5555/seqapi/v1/events/a78ea33299010000-410190f323c58b98" \
  -H "accept: application/json"
```

#### Ответ

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

Скачивает события в файл в указанном формате.

**Авторизация:** ДА

**Тело запроса (application/json):**
- `format` (*enum*, *optional*): Формат экспорта. Одно из `"jsonl"|"csv"` (по умолчанию `"jsonl"`).
- `query` (*string*, *optional*): Поисковый запрос.
- `from` (*string*, *required*): Временная метка начала поиска в `date-time` формате.
- `to` (*string*, *required*): Временная метка окончания поиска в `date-time` формате.
- `limit` (*int*, *required*): Ограничение поиска.
- `offset` (*int*, *optional*): Смещение поиска.
- `fields` (*[]string*, *optional*): Список полей для экспорта (только для `format:csv`, в этом случае список должен быть непустым).

#### Запрос

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

#### Ответ

Данные возвращаются фрагментами (чанками). Конец ответа можно определить по заголовку `Content-Length: 0`.

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

Рассчитывает агрегации на основе событий, удовлетворяющих поисковому запросу.

> Агрегации также могут быть получены с помощью [/search](#post-search).

**Авторизация:** ДА

**Тело запроса (application/json):**
- `query` (*string*, *optional*): Поисковый запрос.
- `from` (*string*, *required*): Временная метка начала поиска в `date-time` формате.
- `to` (*string*, *required*): Временная метка окончания поиска в `date-time` формате.
- `aggregations` (*[]object*, *required*): Список запросов на агрегацию.
  - `agg_func` (*enum*, *optional*): Агрегатная функция. Одно из `"count"|"sum"|"min"|"max"|"avg"|"quantile"|"unique"` (по умолчанию `"count"`).
  - `field` (*string*, *required*): Поле для расчета агрегации.
  - `group_by` (*string*, *optional*): Поле для группировки результатов агрегирования.
  - `quantiles` (*[]int*, *optional*): Список квантилей (только для `agg_func:quantile`, в этом случае список должен быть непустым).

#### Запрос

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

#### Ответ

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

Рассчитывает агрегации в течение различных временных интервалов (также известных как временные ряды) на основе событий, удовлетворяющих поисковому запросу.

**Авторизация:** ДА

**Тело запроса (application/json):**
- `query` (*string*, *optional*): Поисковый запрос.
- `from` (*string*, *required*): Временная метка начала поиска в `date-time` формате.
- `to` (*string*, *required*): Временная метка окончания поиска в `date-time` формате.
- `aggregations` (*[]object*, *required*): Список запросов на агрегацию.
  - `agg_func` (*enum*, *optional*): Агрегатная функция. Одно из `"count"|"sum"|"min"|"max"|"avg"|"quantile"|"unique"` (по умолчанию `"count"`).
  - `interval` (*string*, *required*): Интервал для рассчета сегмента в `duration` формате.
  - `field` (*string*, *required*): Поле для расчета агрегации.
  - `group_by` (*string*, *optional*): Поле для группировки результатов агрегирования.
  - `quantiles` (*[]int*, *optional*): Список квантилей (только для `agg_func:quantile`, в этом случае список должен быть непустым).

#### Запрос

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

#### Ответ

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

Рассчитывает гистограмму на основе событий, удовлетворяющих поисковому запросу.

> Гистограмма также может быть получена с помощью [/search](#post-search).

**Авторизация:** ДА

**Тело запроса (application/json):**
- `query` (*string*, *optional*): Поисковый запрос.
- `from` (*string*, *required*): Временная метка начала поиска в `date-time` формате.
- `to` (*string*, *required*): Временная метка окончания поиска в `date-time` формате.
- `interval` (*string*, *required*): Интервал гистограммы в `duration` формате.

#### Запрос

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

#### Ответ

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

Возвращает список полей, которые будут закреплены в пользовательском интерфейсе. Устанавливается в [разделе конфигурации](./02-configuration.md#seqapi) `handlers.seq_api.pinned_fields`.

**Авторизация:** НЕТ

#### Запрос

```shell
curl -X GET \
  "http://localhost:5555/seqapi/v1/fields/pinned" \
  -H "accept: application/json"
```

#### Ответ

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

Возвращает список ограничений, установленных в [разделе конфигурации](./02-configuration.md#seqapi) `handlers.seq_api`.

**Авторизация:** НЕТ

#### Запрос

```shell
curl -X GET \
  "http://localhost:5555/seqapi/v1/limits" \
  -H "accept: application/json"
```

#### Ответ

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

Возвращает время жизни логов (в секундах) в seq-db.

**Авторизация:** НЕТ

#### Запрос

```shell
curl -X GET \
  "http://localhost:5555/seqapi/v1/logs_lifespan" \
  -H "accept: application/json"
```

#### Ответ

```json
{
  "lifespan": 4923192
}
```

### `GET /status`

Возвращает статус хранилищ seq-db.

**Авторизация:** НЕТ

#### Запрос

```shell
curl -X GET \
  "http://localhost:5555/seqapi/v1/status" \
  -H "accept: application/json"
```

#### Ответ

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