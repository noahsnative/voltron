FROM golang:1.13-alpine3.11 AS builder
RUN apk add --update git make
WORKDIR /app
COPY . .
RUN make linux

FROM alpine:3.9
ENTRYPOINT ["/usr/bin/voltron-injector"]
COPY --from=builder /app/bin/voltron-injector-linux-amd64 /usr/bin/voltron-injector
