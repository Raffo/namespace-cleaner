FROM golang as builder

COPY . /go/src/github.com/Raffo/namespace-cleaner

RUN cd /go/src/github.com/Raffo/namespace-cleaner && make

FROM alpine:3.7

USER nobody

COPY --from=builder /go/src/github.com/Raffo/namespace-cleaner/build/namespace-cleaner-linux-amd64 /usr/local/bin/

ENTRYPOINT ["namespace-cleaner"]