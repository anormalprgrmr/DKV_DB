FROM golang:1.25.2-alpine AS builder

# RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o dkvdb ./cmd/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/dkvdb .

COPY db.db ./

EXPOSE 3000

CMD ["./dkvdb"]
