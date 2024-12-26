FROM golang:1.23-alpine AS builder
WORKDIR /app

COPY . .

RUN go build -o main .

RUN wget https://github.com/docker/compose/releases/download/v2.32.1/docker-compose-linux-x86_64 -O /app/docker-compose

FROM alpine:latest
WORKDIR /app

COPY --from=builder /app/main /app/main
RUN chmod +x /app/main

COPY --from=builder /app/docker-compose /usr/local/bin/docker-compose
RUN chmod +x /usr/local/bin/docker-compose

CMD ["/app/main"]
