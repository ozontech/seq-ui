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

Here is the minimal seq-db config:

`config.seq-db.yaml`
```yaml
cluster:
  hot_stores:
    - seq-db-store:9004

mapping:
  path: auto
```

Here is the minimal docker compose example:
```yaml
services:
  seq-ui:
    image: ghcr.io/ozontech/seq-ui:latest
    volumes:
      - ${PWD}/config/config.example.yaml:/seq-ui/config.yaml
    ports:
      - "5555:5555" # Default HTTP port
      - "5556:5556" # Default gRPC port
      - "5557:5557" # Default debug port
    command: --config config.yaml
    depends_on:
      - seq-db-proxy
  
  seq-db-proxy:
    image: ghcr.io/ozontech/seq-db:latest
    volumes:
      - ${PWD}/config.seq-db.yaml:/seq-db/config.yaml
    ports:
      - "9002:9002" # Default HTTP port
      - "9004:9004" # Default gRPC port
      - "9200:9200" # Default debug port
    command: --mode proxy --config=config.yaml
    depends_on:
      - seq-db-store
  
  seq-db-store:
    image: ghcr.io/ozontech/seq-db:latest
    volumes:
      - ${PWD}/config.seq-db.yaml:/seq-db/config.yaml
    command: --mode store --config=config.yaml
```

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
seq-ui offers many more useful features for working with logs and users. Here's a couple:
- TODO