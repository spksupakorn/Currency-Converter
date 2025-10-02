FROM golang:1.25-alpine AS builder

RUN apk add --no-cache tzdata

ENV TZ=Asia/Bangkok

WORKDIR /app

COPY go.* ./
RUN go mod download

COPY . ./

RUN go build -v -o backend ./cmd/app

FROM alpine:latest

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata \
    && cp /usr/share/zoneinfo/Asia/Bangkok /etc/localtime \
    && echo "Asia/Bangkok" > /etc/timezone

COPY --from=builder /app/backend /app/backend

EXPOSE 8080

CMD ["/bin/sh", "-c", "./backend"]