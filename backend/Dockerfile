FROM golang:1.25-alpine AS builder

WORKDIR /src

RUN apk add --no-cache git

COPY backend/go.mod backend/go.sum ./backend/
WORKDIR /src/backend
RUN go mod download

COPY backend/ .
RUN mkdir -p /out && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/blog-backend .

FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app/backend
COPY --from=builder /out/blog-backend ./blog-backend
RUN mkdir -p uploads

EXPOSE 8080

CMD ["./blog-backend"]
