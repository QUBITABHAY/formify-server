FROM golang:1.21-alpine AS builder

WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/formify-server ./cmd/api

FROM alpine:3.19

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /app/bin/formify-server /app/formify-server

ENV PORT=8080

EXPOSE 8080

USER nobody

CMD ["/app/formify-server"]
