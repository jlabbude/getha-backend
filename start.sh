#!/bin/bash

echo $AVAHI_HOSTNAME > /etc/hostname
hostname $AVAHI_HOSTNAME

mkdir -p /var/run/dbus

if [ -e /run/dbus/pid ] && ! pgrep dbus-daemon > /dev/null; then
    rm -f /run/dbus/pid
fi

if ! pgrep dbus-daemon > /dev/null; then
    dbus-daemon --system --fork
    if [ $? -ne 0 ]; then
        exit 1
    fi
fi

if ! pgrep avahi-daemon > /dev/null; then
    avahi-daemon --debug &
    sleep 2
    if ! pgrep avahi-daemon > /dev/null; then
        exit 1
    fi
fi

ps aux | grep "[a]vahi-daemon"

avahi-publish -s $AVAHI_HOSTNAME _http._tcp 8000 &

exec ./app