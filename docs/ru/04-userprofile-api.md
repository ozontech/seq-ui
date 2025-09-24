# UserProfile API

UserProfile API предоставляет возможность управлять пользователями и их данными.

**Для работы API требуется база данных PostgreSQL и Авторизация, которые должны быть настроены в [конфигурации](./02-configuration.md).**

## HTTP API

**Базовый URL-адрес:** `/userprofile/v1`

Имя пользователя берется из заголовка `Authorization`.

> Вы также можете использовать [swagger-файл](https://github.com/ozontech/seq-ui/blob/main/swagger/swagger.json) для подробного просмотра HTTP API.

### `GET /profile`

Возвращает пользовательские данные:
- Часовой пояс
- Версия пройденного обучения 
- Столбцы таблицы логов (закрепленные столбцы таблицы в пользовательском интерфейсе)

> Если пользователя нет в базе данных, то он будет создан.

**Авторизация:** ДА

#### Запрос

```shell
curl -X GET \
  "http://localhost:5555/userprofile/v1/profile" \
  -H "accept: application/json" \
  -H "Authorization: Bearer <token>"
```

#### Ответ

```json
{
  "timezone": "UTC",
  "onboardingVersion": "{}",
  "log_columns": ["level"]
}
```

### `PATCH /profile`

Обновлят пользовательские данные:
- Часовой пояс
- Версия пройденного обучения 
- Столбцы таблицы логов (закрепленные столбцы таблицы в пользовательском интерфейсе)

**Авторизация:** ДА

**Тело запроса (application/json):**
- `timezone` (*string*, *optional*): Часовой пояс пользователя.
- `onboardingVersion` (*string*, *optional*): Версия пройденного пользователем обучения.
- `log_columns` (*object*, *optional*): Закрепленные пользователем столбцы таблицы логов.
  - `columns` (*[]string*, *required*): Список столбцов.

#### Запрос

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

#### Ответ

```json
{}
```

### `GET /queries/favorite`

Возвращает избранные (сохраненные) поисковые запросы пользователя.

**Авторизация:** ДА

#### Запрос

```shell
curl -X GET \
  "http://localhost:5555/userprofile/v1/queries/favorite" \
  -H "accept: application/json" \
  -H "Authorization: Bearer <token>"
```

#### Ответ

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

Сохраняет поисковый запрос пользователя.

**Авторизация:** ДА

**Тело запроса (application/json):**
- `query` (*string*, *required*): Поисковый запрос.
- `name` (*string*, *optional*): Название поискового запроса.
- `relativeFrom` (*string*, *optional*): Количество секунд относительно текущего времени для вычисления диапазона поиска `from-to`.

#### Запрос

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

#### Ответ

```json
{
  "id": "123"
}
```

### `DELETE /queries/favorite/{id}`

Удаляет определенный избранный поисковый запрос пользователя по его идентификатору.

**Авторизация:** ДА

**Параметры:**
- `id` (*int*, *required*): Уникальный идентификатор избранного запроса.

#### Запрос

```shell
curl -X DELETE \
  "http://localhost:5555/userprofile/v1/queries/favorite/123" \
  -H "accept: application/json" \
  -H "Authorization: Bearer <token>"
```

#### Ответ

```json
{}
```