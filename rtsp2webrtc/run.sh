#!/usr/bin/env bash

cat /data/options.json | jq "
    {server: {http_port: \":8083\"}, streams: [.streams[] | {(.name): {url: .url}}] | add}
    " > /app/config.json

cat /app/config.json

while true; do
    /app/RTSPtoWebRTC
    sleep 5
done