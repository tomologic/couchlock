FROM alpine:3.13
RUN apk --no-cache add ca-certificates
COPY couchlock /usr/bin/
ENTRYPOINT ["/usr/bin/couchlock"]
