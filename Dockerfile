FROM alpine:latest

LABEL maintainer James Rasell<(jamesrasell@gmail.com)> (@jrasell)
LABEL vendor "jrasell"

ENV LEVANT_VERSION 0.2.7

WORKDIR /usr/bin/

RUN buildDeps=' \
                bash \
                wget \
        ' \
        set -x \
        && apk --no-cache add $buildDeps ca-certificates \
        && wget -O levant https://github.com/jrasell/levant/releases/download/${LEVANT_VERSION}/linux-amd64-levant \
        && chmod +x /usr/bin/levant \
        && apk del $buildDeps \
        && echo "Build complete."

CMD ["levant", "--help"]
