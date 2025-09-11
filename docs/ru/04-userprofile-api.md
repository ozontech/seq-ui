# UserProfile API

UserProfile API provides the ability to manage users and their data.

**The API requires a PostgreSQL DB and Authorization to work, which must be specified in [config](https://github.com/ozontech/seq-ui/blob/main/docs/en/02-config.md).**

## HTTP API

**Base URL:** `/userprofile/v1`

The username is taken from the `Authorization` header.

> You can also use [swagger file](https://github.com/ozontech/seq-ui/blob/main/swagger/swagger.json) to view the HTTP API in detail.

### `GET /profile`

Returns user data:
- Timezone
- Onboarding version
- Log columns (Pinned table columns in UI)

> If user doesn't exist in the DB, it will be created.

**Auth:** YES

#### Request

```shell
curl -X GET \
  "http://localhost:5555/userprofile/v1/profile" \
  -H "accept: application/json" \
  -H "Authorization: Bearer <token>"
```

#### Response

```json
{
  "timezone": "UTC",
  "onboardingVersion": "{}",
  "log_columns": ["level"]
}
```

### `PATCH /profile`

Updates user data:
- Timezone
- Onboarding version
- Log columns (Pinned table columns in UI)

**Auth:** YES

**Request Body (application/json):**
- `timezone` (*string*, *optional*): User's timezone.
- `onboardingVersion` (*string*, *optional*): User's onboarding version.
- `log_columns` (*object*, *optional*): User's pinned table columns.
  - `columns` (*[]string*, *required*): List of columns.

#### Request

```shell
curl -X PATCH \
  "http://localhost:5555/userprofile/v1/profile" \
  -H "accept: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '
  {
    "timezone": "UTC",
    "log_columns": {
      "columns": ["source", "error"]
    }
  }'
```

#### Response

```json
{}
```

### `GET /queries/favorite`

Returns user's favorite (saved) search queries.

**Auth:** YES

#### Request

```shell
curl -X GET \
  "http://localhost:5555/userprofile/v1/queries/favorite" \
  -H "accept: application/json" \
  -H "Authorization: Bearer <token>"
```

#### Response

```json
{
  "queries": [
    {
      "id": "1",
      "query": "message:error",
      "name": "error message"
    },
    {
      "id": "2",
      "query": "message:error",
      "relativeFrom": "840"
    },
    {
      "id": "3",
      "query": "level:3",
      "name": "error levels",
      "relativeFrom": "300"
    },
    {
      "id": "4",
      "query": "level:6"
    }
  ]
}
```

### `POST /queries/favorite`

Saves user's search query.

**Auth:** YES

**Request Body (application/json):**
- `query` (*string*, *required*): Search query.
- `name` (*string*, *optional*): Search query name.
- `relativeFrom` (*string*, *optional*): The number of seconds relative to the current time to calculate the `from-to` search range.

#### Request

```shell
curl -X POST \
  "http://localhost:5555/userprofile/v1/queries/favorite" \
  -H "accept: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '
  {
    "query": "message:test",
    "name": "message with test"
  }'
```

#### Response

```json
{
  "id": "123"
}
```

### `DELETE /queries/favorite/{id}`

Deletes a specific favorite (saved) search query by their ID.

**Auth:** YES

**Params:**
- `id` (*int*, *required*): The unique identifier of the favorite query.

#### Request

```shell
curl -X DELETE \
  "http://localhost:5555/userprofile/v1/queries/favorite/123" \
  -H "accept: application/json" \
  -H "Authorization: Bearer <token>"
```

#### Response

```json
{}
```