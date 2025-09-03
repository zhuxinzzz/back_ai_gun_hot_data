FROM golang:1.24-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o back_ai_gun_data .

FROM alpine:latest
WORKDIR /app
COPY --from=build /app/back_ai_gun_data .
CMD ["./back_ai_gun_data"]
