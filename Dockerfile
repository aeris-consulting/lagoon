FROM golang:alpine as builder

RUN apk update && apk upgrade && \
    apk add --no-cache bash git openssh

RUN go get -u github.com/kardianos/govendor

WORKDIR /go/src/lagoon
COPY . .
RUN govendor install
RUN go install -v ./...


FROM alpine
RUN apk update && apk upgrade && \
    apk add --no-cache netcat-openbsd

WORKDIR /lagoon
COPY ./ui/dist ./ui/dist
COPY --from=builder /go/bin/lagoon .

COPY entrypoint.sh /usr/local/bin
RUN chmod +x /usr/local/bin/entrypoint.sh

HEALTHCHECK --start-period=2s --interval=5s --timeout=2s --retries=5 CMD ["nc", "-z", "localhost", "4000"]
ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]

EXPOSE 4000