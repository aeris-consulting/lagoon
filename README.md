# Lagoon


## Project setup
### UI
```
cd ui && npm install
```

### Backend
```
go get -u github.com/kardianos/govendor
govendor install
```

## Run for development
### Start Redis
Put a pre-existing file `appendonly.aof` with data in `docker/data`.
```
cd docker && ./start_redis_single.sh
```

### UI part (Compiles and hot-reloads)
```

cd ui && npm run serve
```

### Backend
```
go build main.go
```

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
cd ui && npm run build
```

### Backend
```
go install -v ./...
```

### Docker image (Build UI first)
```
docker build . -t aerisconsulting/lagoon --no-cache && docker push aerisconsulting/lagoon
```

## UI Misc
### Run the UI tests
```
npm run test
```
### Lints and fixes files
```
npm run lint
```
### Customize configuration
See [Configuration Reference](https://cli.vuejs.org/config/).

## TODO
1. Persist configuration or pass it as base64 YAML file (datasources, listening port)
1. Manage read-only datasources
1. Document the API
1. Edit a content
1. Add a root or child entrypoint (manage the type of entry point to be managed as properties from the datasource)
1. Visualize content as formatted and pretty-printed JSON / YAML
1. Manage multitab, pining tabs (close all and close unpins) to display content
1. Add a filter for the content (useful for long sets)
1. Copy visible content to the clipboard
1. Support Kafka and streams
1. Manage templates with placeholders to create content
1. Extend basic features: 

  * A "search" Tab, in addition to the "explore" one to navigate in all the 
entrypoints and find the ones matching a simple query, like: `SET(my-key:*) HAVING length > 2` or `HASH(my-key:*) HAVING FIELD foo = bar`
  * A visualization tab to see the evolution of data over time
  * A processing tab (map, reduce)


## Resources
* Web
  * https://gin-gonic.com/docs/examples/
  * http://arlimus.github.io/articles/gin.and.gorilla/
  * https://gin-gonic.com/docs/examples/graceful-restart-or-stop/
* Backend
  * https://github.com/kardianos/govendor
  * https://gin-gonic.com/docs/examples/
  * https://github.com/nathantsoi/vue-native-websocket
* Redis
  * https://github.com/go-redis/redis/blob/master/example_test.go
  * https://godoc.org/github.com/go-redis/redis
  * https://redis.io/commands
* Consume web-socket in shell
  * https://github.com/websockets/wscat