FROM golang:1.13-alpine3.10 AS BUILD-ENV

ARG GOOS_VAL 
ARG GOARCH_VAL

# Install deps
RUN apk update && apk upgrade && \
    apk add --no-cache make git

WORKDIR /app

# Download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY pkg ./pkg
COPY cmd ./cmd
COPY Makefile ./

# Build binary
RUN make

ENTRYPOINT ["kyaml2go"]
