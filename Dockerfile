FROM golang:1.14 AS BUILD

LABEL maintainer="Gabriele Paggi"

COPY . /go/src/github.com/hashicorp/levant

WORKDIR /go/src/github.com/hashicorp/levant

RUN make build

FROM alpine:latest

COPY --from=BUILD /go/src/github.com/hashicorp/levant/bin/levant /usr/local/bin/levant

ENTRYPOINT ["/usr/local/bin/levant"]
