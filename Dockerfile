FROM alpine:3.18
COPY vector-db /usr/bin/vector-db
ENTRYPOINT ["/usr/bin/vector-db"]