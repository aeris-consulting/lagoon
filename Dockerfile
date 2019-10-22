# Builder: go-builder
FROM golang:alpine as go-builder

RUN apk update && apk upgrade && \
    apk add --no-cache bash git openssh gcc alpine-sdk

ENV GOPROXY=https://proxy.golang.org

WORKDIR /go/src/lagoon
COPY ./src/lagoon .
RUN go get -v -t -d ./...
RUN go build -v ./...
RUN go install -v ./...

# Builder: node-builder
FROM node as node-builder

RUN apt-get -y update \
	&& apt-get install -y git

RUN yarn global add @vue/cli -g
RUN yarn global add  @vue/cli-service -g

WORKDIR /ui
COPY ./ui .
RUN yarn install
RUN yarn build

# Final image
FROM alpine
RUN apk update && apk upgrade && \
    apk add --no-cache netcat-openbsd bash

WORKDIR /lagoon
COPY --from=node-builder /ui/dist ./ui/dist
COPY --from=go-builder /go/bin/lagoon .

COPY entrypoint.sh /usr/local/bin
RUN chmod +x /usr/local/bin/entrypoint.sh

HEALTHCHECK --start-period=2s --interval=5s --timeout=2s --retries=5 CMD ["nc", "-z", "localhost", "4000"]
ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]

EXPOSE 4000