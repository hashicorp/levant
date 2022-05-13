# This Dockerfile contains multiple targets.
# Use 'docker build --target=<name> .' to build one.

# ===================================
#   Non-release images.
# ===================================

# devbuild compiles the binary
# -----------------------------------
FROM golang:latest AS devbuild

# Disable CGO to make sure we build static binaries
ENV CGO_ENABLED=0

# Escape the GOPATH
WORKDIR /build
COPY . ./
RUN go build -o levant .


# dev runs the binary from devbuild
# -----------------------------------
FROM alpine:latest AS dev
COPY --from=devbuild /build/levant /bin/

ENTRYPOINT ["/bin/levant"]
CMD ["-v"]


# ===================================
#   Release images.
# ===================================

FROM alpine:latest AS release

ARG PRODUCT_NAME=levant
ARG PRODUCT_VERSION
ARG PRODUCT_REVISION
# TARGETARCH and TARGETOS are set automatically when --platform is provided.
ARG TARGETOS TARGETARCH

LABEL maintainer="Nomad Team <nomad@hashicorp.com>"
LABEL version=${PRODUCT_VERSION}
LABEL revision=${PRODUCT_REVISION}

COPY dist/$TARGETOS/$TARGETARCH/levant /bin/

# Create a non-root user to run the software.
RUN addgroup $PRODUCT_NAME && \
    adduser -S -G $PRODUCT_NAME $PRODUCT_NAME

USER $PRODUCT_NAME
ENTRYPOINT ["/bin/levant"]
CMD ["-v"]

# ===================================
#   Set default target to 'dev'.
# ===================================
FROM dev
