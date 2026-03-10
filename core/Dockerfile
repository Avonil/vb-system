FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
# Збираємо статичний бінарник
RUN CGO_ENABLED=0 GOOS=linux go build -o veris-core main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/veris-core .

EXPOSE 9000
CMD ["./veris-core"]