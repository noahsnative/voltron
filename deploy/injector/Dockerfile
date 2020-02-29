FROM golang:1.13-alpine3.11 AS builder
RUN apk add --update git make
WORKDIR /app
ADD . .
RUN make linux

FROM alpine:3.9
ENTRYPOINT ["/usr/bin/injector"]
COPY --from=builder /app/bin/injector-linux-amd64 /usr/bin/injector
