FROM golang:alpine as builder

RUN apk update && apk upgrade && \
    apk add --no-cache bash git openssh

RUN go get -u github.com/kardianos/govendor

WORKDIR /go/src/lagoon
COPY . .

RUN govendor list \
    govendor install

RUN go install -v ./...

RUN which lagoon

FROM alpine

WORKDIR /lagoon
COPY ./ui/dist ./ui/dist
COPY --from=builder /go/bin/lagoon .

EXPOSE 4000
ENTRYPOINT ["./lagoon"]
