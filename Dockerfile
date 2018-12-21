FROM alpine:3.7

ADD build/namespace-cleaner-linux-amd64 /namespace-cleaner

ENTRYPOINT ["/namespace-cleaner"]