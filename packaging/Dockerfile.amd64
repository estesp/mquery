# This requires a static binary for mquery; see `make static`
FROM alpine:3.6

RUN apk update && apk add ca-certificates

COPY mquery-linux-amd64 /bin/mquery

ENTRYPOINT ["/bin/mquery"]
