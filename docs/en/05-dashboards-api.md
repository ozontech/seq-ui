# Dashboards API

Dashboards API provides the ability to manage dashboards.

The dashboard consists of:
- search query
- search interval
- widgets: aggregations and a histogram
- list of pinned table columns

**The API requires a PostgreSQL DB and Authorization to work, which must be specified in [config](./02-configuration.md).**

## HTTP API

**Base URL:** `/dashboards/v1`

The dashboard owner is taken from the `Authorization` header.

> You can also use [swagger file](https://github.com/ozontech/seq-ui/blob/main/swagger/swagger.json) to view the HTTP API in detail.

### `POST /`

Creates dashboard.

**Auth:** YES

**Request Body (application/json):**
- `name` (*string*, *required*): Dashboard name.
- `meta` (*string*, *required*): Dashboard metadata in `json`-format that is used in frontend app.

#### Request

```shell
curl -X POST \
  "http://localhost:5555/dashboards/v1/" \
  -H "accept: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '
  {
    "name": "my dashboard",
    "meta": "{\"histogram\":false,\"aggregations\":[{\"fn\":\"count\",\"field\":\"level\"}],\"query\":\"_exists_:level\",\"columns\":[\"level\"]}"
  }'
```

#### Response

```json
{
  "uuid": "066333fc-0317-7000-b1b1-e2ceaa140af1"
}
```

### `POST /all`

Returns list of dashboards of all users.

**Auth:** YES

**Request Body (application/json):**
- `limit` (*int*, *required*): Limit of the returned list.
- `offset` (*int*, *optional*): Offset from beginning of the list.

#### Request

```shell
curl -X POST \
  "http://localhost:5555/dashboards/v1/all" \
  -H "accept: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '
  {
    "limit": 3,
    "offset": 0
  }'
```

#### Response

```json
{
  "dashboards": [
    {
      "uuid": "066b2360-9037-7000-9382-9b33a690bf17",
      "name": "old",
      "owner_name": "bobrov11"
    },
    {
      "uuid": "066b236b-a06c-7000-82ff-62bd329b5123",
      "name": "my dashboard",
      "owner_name": "ivanovivan"
    },
    {
      "uuid": "066b235e-52a4-7000-9fee-759d4b0b152c",
      "name": "test",
      "owner_name": "ivanovivan"
    }
  ]
}
```

### `POST /my`

Returns list of dashboards of the current user.

**Auth:** YES

**Request Body (application/json):**
- `limit` (*int*, *required*): Limit of the returned list.
- `offset` (*int*, *optional*): Offset from beginning of the list.

#### Request

```shell
curl -X POST \
  "http://localhost:5555/dashboards/v1/my" \
  -H "accept: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '
  {
    "limit": 2,
    "offset": 0
  }'
```

#### Response

```json
{
  "dashboards": [
    {
      "uuid": "066b236b-a06c-7000-82ff-62bd329b5123",
      "name": "my dashboard"
    },
    {
      "uuid": "066b235e-52a4-7000-9fee-759d4b0b152c",
      "name": "test"
    }
  ]
}
```

### `POST /search`

Returns list of dashboards that satisfy the search query. The search is performed by the name of the dashboard.

**Auth:** YES

**Request Body (application/json):**
- `query` (*string*, *required*): Search query.
- `limit` (*int*, *required*): Limit of the returned list.
- `offset` (*int*, *optional*): Offset from beginning of the list.
- `filter` (*object*, *optional*): Search filter.
  - `owner_name` (*string*, *optional*): Filter by owner name.

#### Request

```shell
curl -X POST \
  "http://localhost:5555/dashboards/v1/search" \
  -H "accept: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '
  {
    "query": "test",
    "limit": 2,
    "offset": 0
  }'
```

#### Response

```json
{
  "dashboards": [
    {
      "uuid": "066b236b-a06c-7000-82ff-62bd329b5123",
      "name": "my test dashboard",
      "owner_name": "ivanivanov"
    },
    {
      "uuid": "066b235e-52a4-7000-9fee-759d4b0b152c",
      "name": "123test123",
      "owner_name": "bobrov11"
    }
  ]
}
```

### `GET /{uuid}`

Retrieves a specific dashboard by their ID.

**Auth:** YES

**Params:**
- `uuid` (*string*, *required*): The unique identifier of the dashboard.

#### Request

```shell
curl -X GET \
  "http://localhost:5555/dashboards/v1/066333fc-0317-7000-b1b1-e2ceaa140af1" \
  -H "accept: application/json" \
  -H "Authorization: Bearer <token>"
```

#### Response

```json
{
  "name": "my dashboard",
  "meta": "{\"histogram\":false,\"aggregations\":[{\"fn\":\"count\",\"field\":\"level\"}],\"query\":\"_exists_:level\",\"columns\":[\"level\"]}",
  "owner_name": "ivanivanov"
}
```

### `PATCH /{uuid}`

Updates a specific dashboard by their ID.

**Auth:** YES

**Params:**
- `uuid` (*string*, *required*): The unique identifier of the dashboard.

**Request Body (application/json):**
- `name` (*string*, *optional*): Dashboard name.
- `meta` (*string*, *optional*): Dashboard metadata in `json`-format that is used in frontend app.

#### Request

```shell
curl -X PATCH \
  "http://localhost:5555/dashboards/v1/066333fc-0317-7000-b1b1-e2ceaa140af1" \
  -H "accept: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '
  {
    "name": "new dashboard name"
  }'
```

#### Response

```json
{}
```

### `DELETE /{uuid}`

Deletes a specific dashboard by their ID.

**Auth:** YES

**Params:**
- `uuid` (*string*, *required*): The unique identifier of the dashboard.

#### Request

```shell
curl -X DELETE \
  "http://localhost:5555/dashboards/v1/066333fc-0317-7000-b1b1-e2ceaa140af1" \
  -H "accept: application/json" \
  -H "Authorization: Bearer <token>"
```

#### Response

```json
{}
```