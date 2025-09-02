FROM golang:1.21-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download || true

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o 0 main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

RUN addgroup -g 1001 -S appgroup &&     adduser -u 1001 -S appuser -G appgroup

WORKDIR /app

COPY --from=builder /app/0 .

RUN chown -R appuser:appgroup /app

USER appuser

EXPOSE 8114

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3     CMD wget --no-verbose --tries=1 --spider http://localhost:8114/health || exit 1

CMD ["./0", "-port=8114"]
