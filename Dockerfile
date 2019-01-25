FROM golang as builder

COPY . /go/src/github.com/Raffo/namespace-cleaner

RUN cd /go/src/github.com/Raffo/namespace-cleaner && make build

FROM alpine:3.7

USER nobody

COPY --from=builder /go/src/github.com/Raffo/namespace-cleaner/build/namespace-cleaner-linux-amd64 /namespace-cleaner

RUN ls -la
# COPY build/namespace-cleaner-linux-amd64 /namespace-cleaner

ENTRYPOINT ["/namespace-cleaner"]