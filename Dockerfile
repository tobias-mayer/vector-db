FROM alpine:3.17
COPY vector-db /usr/bin/vector-db
ENTRYPOINT ["/usr/bin/vector-db"]