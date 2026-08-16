FROM golang:1.25.6-alpine AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /wallet ./cmd/app

FROM alpine:3.22

WORKDIR /app
COPY --from=build /wallet /app/wallet

EXPOSE 8080
ENTRYPOINT ["/app/wallet"]
