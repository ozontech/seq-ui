# Dashboards API

Dashboards API предоставляет возможность управлять дашбордами.

Дашборд состоит из:
- поискового запроса
- интервала поиска
- виджетов: агрегаций и гистограммы
- списка закрепленных столбцов таблицы логов

**Для работы API требуется база данных PostgreSQL и Авторизация, которые должны быть настроены в [конфигурации](./02-configuration.md).**

## HTTP API

**Базовый URL-адрес:** `/dashboards/v1`

Владелец дашборда берется из заголовка `Authorization`.

> Вы также можете использовать [swagger-файл](https://github.com/ozontech/seq-ui/blob/main/swagger/swagger.json) для подробного просмотра HTTP API.

### `POST /`

Создает дашборд.

**Авторизация:** ДА

**Тело запроса (application/json):**
- `name` (*string*, *required*): Название дашборда.
- `meta` (*string*, *required*): Метаданные дашборда в формате `json`, которые используются frontend-приложением.

#### Запрос

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

#### Ответ

```json
{
  "uuid": "066333fc-0317-7000-b1b1-e2ceaa140af1"
}
```

### `POST /all`

Возвращает список дашбордов всех пользователей.

**Авторизация:** ДА

**Тело запроса (application/json):**
- `limit` (*int*, *required*): Ограничение размера возвращаемого списка.
- `offset` (*int*, *optional*): Смещение от начала списка.

#### Запрос

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

#### Ответ

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

Возвращает список дашбордов текущего пользователя.

**Авторизация:** ДА

**Тело запроса (application/json):**
- `limit` (*int*, *required*): Ограничение размера возвращаемого списка.
- `offset` (*int*, *optional*): Смещение от начала списка.

#### Запрос

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

#### Ответ

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

Возвращает список дашбордов, удовлетворяющих поисковому запросу. Поиск выполняется по названию дашборда.

**Авторизация:** ДА

**Тело запроса (application/json):**
- `query` (*string*, *required*): Поисковый запрос.
- `limit` (*int*, *required*): Ограничение размера возвращаемого списка.
- `offset` (*int*, *optional*): Смещение от начала списка.
- `filter` (*object*, *optional*): Поисковый фильтр.
  - `owner_name` (*string*, *optional*): Фильтрация по владельцу.

#### Запрос

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

#### Ответ

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

Извлекает определенный дашборд по его идентификатору.

**Авторизация:** ДА

**Параметры:**
- `uuid` (*string*, *required*): Уникальный идентификатор дашборда.

#### Запрос

```shell
curl -X GET \
  "http://localhost:5555/dashboards/v1/066333fc-0317-7000-b1b1-e2ceaa140af1" \
  -H "accept: application/json" \
  -H "Authorization: Bearer <token>"
```

#### Ответ

```json
{
  "name": "my dashboard",
  "meta": "{\"histogram\":false,\"aggregations\":[{\"fn\":\"count\",\"field\":\"level\"}],\"query\":\"_exists_:level\",\"columns\":[\"level\"]}",
  "owner_name": "ivanivanov"
}
```

### `PATCH /{uuid}`

Обновляет определенный дашборд по его идентификатору.

**Авторизация:** ДА

**Параметры:**
- `uuid` (*string*, *required*): Уникальный идентификатор дашборда.

**Тело запроса (application/json):**
- `name` (*string*, *optional*): Название дашборда.
- `meta` (*string*, *optional*): Метаданные дашборда в формате `json`, которые используются frontend-приложением.

#### Запрос

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

#### Ответ

```json
{}
```

### `DELETE /{uuid}`

Удаляет определенный дашборд по его идентификатору.

**Авторизация:** ДА

**Параметры:**
- `uuid` (*string*, *required*): Уникальный идентификатор дашборда.

#### Запрос

```shell
curl -X DELETE \
  "http://localhost:5555/dashboards/v1/066333fc-0317-7000-b1b1-e2ceaa140af1" \
  -H "accept: application/json" \
  -H "Authorization: Bearer <token>"
```

#### Ответ

```json
{}
```