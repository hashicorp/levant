FROM golang:alpine AS builder

RUN buildDeps=' \
                make \
                git \
        ' \
        set -x \
        && apk --no-cache add $buildDeps \
        && mkdir -p /go/src/github.com/jrasell/levant

WORKDIR /go/src/github.com/jrasell/levant

COPY . /go/src/github.com/jrasell/levant

RUN \
        make tools && \
        make build

FROM alpine:latest AS app

LABEL maintainer James Rasell<(jamesrasell@gmail.com)> (@jrasell)
LABEL vendor "jrasell"

WORKDIR /usr/bin/

COPY --from=builder /go/src/github.com/jrasell/levant/levant-local /usr/bin/levant

RUN \
        apk --no-cache add \
        ca-certificates \
        && chmod +x /usr/bin/levant \
        && echo "Build complete."

CMD ["levant", "--help"]