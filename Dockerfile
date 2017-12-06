FROM golang:alpine as builder

WORKDIR /go/src/github.com/jrasell/levant
COPY . .
RUN apk add --no-cache make \
    && make install

FROM alpine:latest

RUN apk add --no-cache ca-certificates

COPY --from=builder /go/bin/levant /usr/bin/levant
CMD ["levant", "--help"]
