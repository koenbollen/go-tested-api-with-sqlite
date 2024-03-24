FROM golang:1.22.1-alpine@sha256:0466223b8544fb7d4ff04748acc4d75a608234bf4e79563bff208d2060c0dd79 AS builder

WORKDIR /go-tested-api-with-sqlite

RUN apk --no-cache add ca-certificates git

COPY go* ./
RUN go mod download

COPY ./ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o api -a -ldflags '-extldflags "-static"' cmd/api/main.go

FROM scratch
EXPOSE 8080
ENTRYPOINT ["/api"]

COPY --from=builder /go-tested-api-with-sqlite/api /api
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
