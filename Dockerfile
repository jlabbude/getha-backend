FROM golang:1.23 as builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o app .

FROM ubuntu:latest
WORKDIR /app
COPY --from=builder /app/app .
EXPOSE 8000
RUN mkdir -p /app/aparelhos/image /app/aparelhos/video /app/aparelhos/manual
CMD ["./app"]