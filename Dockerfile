FROM golang:1.23 as builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o app .

FROM ubuntu:latest
WORKDIR /app
RUN apt-get update && apt-get install -y \
    avahi-daemon \
    avahi-utils \
    libnss-mdns \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/app .

RUN mkdir -p /app/aparelhos/image /app/aparelhos/video /app/aparelhos/manual

COPY avahi-daemon.conf /etc/avahi/avahi-daemon.conf
COPY start.sh /app/start.sh
RUN chmod +x /app/start.sh
RUN mkdir -p /etc/avahi/services
COPY getha.service /etc/avahi/services/

EXPOSE 8000
EXPOSE 5353/udp

ENV AVAHI_HOSTNAME=getha-backend

CMD ["/app/start.sh"]
