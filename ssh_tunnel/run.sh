#!/usr/bin/env bashio
set -e

# https://github.com/hassio-addons/bashio

bashio::log.info "Public key:"

cat /root/.ssh/id_rsa.pub

CMD='-o StrictHostKeyChecking=no -N'

if bashio::config.true 'tunnel_http'; then
    CMD="${CMD} -R 80:localhost:80"
fi

if bashio::config.true 'tunnel_https'; then
    CMD="${CMD} -R 443:localhost:443"
fi

if ! bashio::config.equals 'socks_port' 0; then
    CMD="${CMD} -D *:$(bashio::config 'socks_port')"
fi

if ! bashio::config.is_empty 'advanced'; then
    CMD="${CMD} $(bashio::config 'advanced')"
fi

if ! bashio::config.equals 'ssh_port' 0; then
    CMD="${CMD} -p $(bashio::config 'ssh_port')"
fi

CMD="${CMD} $(bashio::config 'ssh_user')@$(bashio::config 'ssh_host')"

while true
do
    bashio::log.info "run tunnel: ssh ${CMD}"
    sh -c "ssh ${CMD}"
    sleep 1
done