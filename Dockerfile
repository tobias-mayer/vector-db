FROM alpine:3.19
COPY vector-db /usr/bin/vector-db
ENTRYPOINT ["/usr/bin/vector-db"]