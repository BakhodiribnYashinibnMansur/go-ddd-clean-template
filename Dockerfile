# Step 1: Modules caching
FROM golang:1.26.0-alpine3.21 AS modules

COPY go.mod go.sum /modules/

WORKDIR /modules

RUN go mod download

# Step 2: Install protoc and generate protobuf
FROM golang:1.26.0-alpine3.21 AS proto-generator

COPY --from=modules /go/pkg /go/pkg
COPY . /app

WORKDIR /app

# Install protoc
RUN apk add --no-cache protobuf

# Install Go protobuf plugins
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest && \
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Generate protobuf files
RUN mkdir -p v1/genproto && \
    protoc --go_out=v1/genproto \
           --go_opt=paths=source_relative \
           --go-grpc_out=v1/genproto \
           --go-grpc_opt=paths=source_relative \
           docs/proto/v1/*.proto

# Step 3: Builder
FROM golang:1.26.0-alpine3.21 AS builder

COPY --from=modules /go/pkg /go/pkg
COPY --from=proto-generator /app /app

WORKDIR /app

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -tags migrate -o /bin/app ./cmd/app

# Step 4: Final
FROM gcr.io/distroless/static-debian12

COPY --from=builder /app/config /config
COPY --from=builder /app/migrations /migrations
COPY --from=builder /bin/app /app
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

CMD ["/app"]
