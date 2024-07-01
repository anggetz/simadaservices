FROM golang:1.21.2-alpine AS builder

WORKDIR /app
COPY . .

RUN go build -o ./executable -v ./cmd/service-scheduler/main.go

FROM alpine:3.16

COPY --from=builder /app/.env .
COPY --from=builder /app/executable .
COPY --from=builder /app/storage/logs /storage/logs

CMD ./executable