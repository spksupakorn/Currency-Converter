FROM golang:1.25-alpine AS builder

RUN apk add --no-cache tzdata

ENV TZ=Asia/Bangkok

WORKDIR /app

COPY go.* ./
RUN go mod download

COPY . ./

RUN go build -v -o migration ./database/migrations
RUN go build -v -o backend ./cmd/app

FROM alpine:latest

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata \
    && cp /usr/share/zoneinfo/Asia/Bangkok /etc/localtime \
    && echo "Asia/Bangkok" > /etc/timezone

COPY --from=builder /app/backend /app/backend
COPY --from=builder /app/migration /app/migration
COPY --from=builder /app/uploads /app/uploads

EXPOSE 8080

# Run migration and start the server
CMD ["/bin/sh", "-c", "./migration && ./backend"]