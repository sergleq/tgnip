# --- сборка ---
FROM golang:1.23-alpine AS build
WORKDIR /app
RUN apk add --no-cache ca-certificates git
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /app/bot .

# --- запуск ---
FROM alpine:3.20
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=build /app/bot /app/bot
EXPOSE 8080
ENTRYPOINT ["/app/bot"]
    
    