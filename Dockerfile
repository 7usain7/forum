FROM golang:1.24-alpine AS builder

LABEL version="1.0.0"
LABEL description="Go web application for a forum, similar to Reddit."
LABEL org.opencontainers.image.source="https://learn.reboot01.com/git/mohkadhem/forum"

RUN apk add --no-cache gcc musl-dev

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=1 go build -o main .

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/main ./
COPY --from=builder /app/web ./web
COPY --from=builder /app/database ./database

EXPOSE 8080

CMD [ "./main" ]