# Lagoon


## Project setup
### UI
[Yarn](https://yarnpkg.com) is recommended for front-end development.
```
cd ui && yarn
```

### Backend
```
export GOPATH=$(pwd)
cd src/lagoon
go get -v -t -d ./...
```

### Generating backend mocks
```
export GOPATH=$(pwd)
cd src/lagoon
mockgen -source=datasource/datasource.go -destination=datasource/datasource_mock.go -package=datasource
```


## Run for development
### Start Redis
Put a pre-existing file `appendonly.aof` with data in `docker/data`.
```
cd docker && ./start_redis_single.sh
```

### UI part (Compiles and hot-reloads)
```
cd ui && yarn serve
```

### Backend
```
export GOPATH=$(pwd)
cd src/lagoon
go build ./...
go run main.go
```

Optionally, you can pass configuration in different way. Using a base64 string or a path to a configuration file.

For base64 configuration, pass it like:
```
go run main.go -b=cG9ydDogMjAwMAoKZGF0YXNvdXJjZXM6Ci0gdXVpZDogYjk3MzYyMjQtMWFiMy00OWQ4LWE1OWQtZWYyNWFlNzA5NDg3CiAgdmVuZG9yOiByZWRpcwogIG5hbWU6IFJlZGlzIC0gQ2x1c3RlcgogIGJvb3RzdHJhcDogY2x1c3RlcjovLzEyNy4wLjAuMToxMzAwMSwxMjcuMC4wLjE6MTMwMDIsMTI3LjAuMC4xOjEzMDAzLDEyNy4wLjAuMToxMzAwNCwxMjcuMC4wLjE6MTMwMDUsMTI3LjAuMC4xOjEzMDA2CiAgY29uZmlndXJhdGlvbjoKICAgIHJlYWRUaW1lb3V0OiAzMAogICAgd3JpdGVUaW1lb3V0OiAzMAogICAgbWF4Q29ubkFnZTogMzAKICAgIG1pbklkbGVDb25uczogMTAKLSB1dWlkOiA2YTRhYmRkOS0zNWJlLTRkNGEtYWU0Ni0xZjFjNzBhN2FkMjYKICB2ZW5kb3I6IHJlZGlzCiAgbmFtZTogUmVkaXMgLSBTaW5nbGUKICBib290c3RyYXA6IHJlZGlzOi8vbG9jYWxob3N0OjYzNzkKICAgIAogICAg
```

This is equivalent to the following configuration:
```yaml
port: 2000

datasources:
- id: local-cluster
  vendor: redis
  name: Local Cluster
  bootstrap: cluster://127.0.0.1:13001,127.0.0.1:13002,127.0.0.1:13003,127.0.0.1:13004,127.0.0.1:13005,127.0.0.1:13006
  readonly: true
  configuration:
    readTimeout: 30
    writeTimeout: 30
    maxConnAge: 30
    minIdleConns: 10
- id: single
  vendor: redis
  name: Single
  bootstrap: redis://localhost:6379
  readonly: false
```

Otherwise the configuration will be loaded from the file set with parameter `-c` (default `lagoon.yml`) if it exists.

### Declare the local database
```
curl -X PUT \
  http://localhost:4000/datasource \
  -H 'Content-Type: application/json;charset=UTF-8' \
  -H 'Postman-Token: 70e64a69-b2f8-440d-8001-b141c3d657be' \
  -H 'cache-control: no-cache' \
  -d '{"vendor":"redis","name":"local", "bootstrap":"redis://localhost:6379"}'
```

## Build 
### UI 
```
cd ui && yarn build
```

### Backend
```
go install -v ./...
```

### Docker image (No need to build anything first)
```
docker build . -t aerisconsulting/lagoon && docker push aerisconsulting/lagoon
```

## UI Misc
### Run the UI tests
Running unit test with npm seems problematic, I am not able to fix the issue, therefore it's recommended to use [yarn](https://yarnpkg.com) [github issue](https://github.com/vuejs/vue-cli/issues/1879#issuecomment-412300256)
```
yarn test:unit
```
### Lints and fixes files
```
yarn lint
```
### Customize configuration
See [Configuration Reference](https://cli.vuejs.org/config/).

## TODO

1. Document the API
1. Edit a content
1. Add a root or child entrypoint (manage the type of entry point to be managed as properties from the datasource)
1. Manage multitab, pining tabs (close all and close unpins) to display content
1. Add a filter for the content (useful for long sets)
1. Support Kafka and streams
1. Manage templates with placeholders to create content
1. Extend basic features: 

  * A "search" Tab, in addition to the "explore" one to navigate in all the 
entrypoints and find the ones matching a simple query, like: `SET(my-key:*) HAVING length > 2` or `HASH(my-key:*) HAVING FIELD foo = bar`
  * A visualization tab to see the evolution of data over time
  * A processing tab (map, reduce)


## Resources
* Redis
  * https://github.com/go-redis/redis/blob/master/example_test.go
  * https://godoc.org/github.com/go-redis/redis
  * https://redis.io/commands
* Consume web-socket in shell
  * https://github.com/websockets/wscat