FROM golang:1.17-alpine as builder
WORKDIR /go/src/couchlock
RUN apk --no-cache add git
COPY *.go go.mod ./
RUN go build -v

FROM alpine:3.13
RUN apk --no-cache add ca-certificates
COPY --from=builder /go/src/couchlock/couchlock /usr/bin/

ENTRYPOINT ["couchlock"]
