# Quickstart

Welcome to the seq-ui quickstart guide! In just a few minutes, you'll learn how to:
- Quickly create a seq-ui instance
- Get [seq-db](https://github.com/ozontech/seq-db) indexed fields and search some events using `/seqapi`

## Running seq-ui

Before launch you need to create config file based on the [example config](https://github.com/ozontech/seq-ui/tree/main/config/config.example.yaml) or use it as-is.

seq-ui can be quickly launched in a docker container. Pull seq-ui image from Docker hub and create a container:
```shell
docker run --rm \
  --name seq-ui \
  -p 5555:5555 \
  -p 5556:5556 \
  -p 5557:5557 \
  -v "$(pwd)"/config/config.example.yaml:/seq-ui/config.yaml \
  -it ghcr.io/ozontech/seq-ui:latest --config=config.yaml
```

## Running seq-ui with seq-db

Before next steps, we need to setup seq-db. See [seq-db quickstart](https://github.com/ozontech/seq-db/blob/main/docs/en/01-quickstart.md) for details.

## Get seq-db indexed fields

seq-db doesn't index any fields from the ingested data by default. Instead, indexing is controlled through a special file called the mapping file. See [index types](https://github.com/ozontech/seq-db/blob/main/docs/en/03-index-types.md) for details.

The fields can be obtained using `/seqapi/v1/fields`:
```shell
curl -X GET \
  "http://localhost:5555/seqapi/v1/fields" \
  -H "accept: application/json"
```

## Search for events

Search last 10 events with simple query that filters logs by `message` or `level` fields using `/seqapi/v1/search`:
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

## What's next

seq-ui offers many more useful features for working with logs and users:
- [Seq API](https://github.com/ozontech/seq-ui/blob/main/docs/en/03-seq-api.md) provides access to logs, aggregations and histogram
- [UserProfile API](https://github.com/ozontech/seq-ui/blob/main/docs/en/04-userprofile-api.md) provides the ability to manage users and their data
- [Dashboards API](https://github.com/ozontech/seq-ui/blob/main/docs/en/05-dashboards-api.md) provides the ability to combine a search query, aggregations and a histogram in a dashboard and save it to DB