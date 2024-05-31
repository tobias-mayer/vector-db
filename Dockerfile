FROM alpine:3.20
COPY vector-db /usr/bin/vector-db
ENTRYPOINT ["/usr/bin/vector-db"]