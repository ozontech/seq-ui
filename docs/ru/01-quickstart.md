# Быстрый запуск

Добро пожаловать в руководство по быстрому запуску seq-ui! Всего через несколько минут вы узнаете, как:
- Быстро запустить seq-ui
- Получить список индексируемых полей [seq-db](https://github.com/ozontech/seq-db)
- Искать события (логи) используя `/seqapi`

## Запуск seq-ui

Перед запуском вам нужно создать файл конфигурации, взяв за основу [пример](https://github.com/ozontech/seq-ui/tree/main/config/config.example.yaml), или использовать его без изменений.

seq-ui можно быстро запустить в docker-контейнере. Следующая команда скачает образ seq-ui из GitHub Container Registry (GHCR) и запустит контейнер:
```shell
docker run --rm \
  --name seq-ui \
  -p 5555:5555 \
  -p 5556:5556 \
  -p 5557:5557 \
  -v "$(pwd)"/config/config.example.yaml:/seq-ui/config.yaml \
  -it ghcr.io/ozontech/seq-ui:latest --config=config.yaml
```

## Запуск seq-ui вместе с seq-db

Перед тем как перейти к следующим шагам, необходимо запустить seq-db. Обратитесь к [Быстрому запуску seq-db](https://github.com/ozontech/seq-db/blob/main/docs/ru/01-quickstart.md) за детальной информацией.

## Список индексируемых полей seq-db

По умолчанию, seq-db не индексирует поля из записываемых данных, вместо этого есть специальный файл маппинга, в котором указаны индексируемые поля и используемые типы индексов. Для более детальной информации смотрите [Типы индексов](https://github.com/ozontech/seq-db/blob/main/docs/ru/03-index-types.md).

Спиоск полей может быть получен, используя `/seqapi/v1/fields`:
```shell
curl -X GET \
  "http://localhost:5555/seqapi/v1/fields" \
  -H "accept: application/json"
```

## Поиск событий

Выполним поиск последних `10` событий с простым поисковым запросом, фильтрующим события по полю `message` или `level`, используя `/seqapi/v1/search`:
```shell
curl -X POST \
  "http://localhost:5555/seqapi/v1/search" \
  -H "accept: application/json" \
  -H "Content-Type: application/json" \
  -d '
  {
    "query": "message:error or level:3",
    "from": "2025-09-01T07:00:00Z",
    "to": "2025-09-01T08:00:00Z",
    "limit": 10,
    "offset": 0
  }'
```

## Что дальше

seq-ui имеет множество других функций для работы с логами и пользователями:
- [Seq API](./03-seq-api.md) предоставляет доступ к логам, агрегациям и гистограмме
- [UserProfile API](./04-userprofile-api.md) предоставляет возможность управлять пользователями и их данными
- [Dashboards API](./05-dashboards-api.md) предоставляет возможность объединять поисковый запрос, аггрегации и гистограмму в дашборд, сохраняя его в базе данных