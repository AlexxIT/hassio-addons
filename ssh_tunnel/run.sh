#!/usr/bin/with-contenv bashio

set -e

# https://github.com/hassio-addons/bashio

bashio::log.info "Public key:"

cat /root/.ssh/id_rsa.pub

BASE="ssh -o StrictHostKeyChecking=no"
TUN="-o ExitOnForwardFailure=yes -o ServerAliveInterval=30 -N"

if bashio::config.true 'tunnel_http'; then
    # -R remote_socket:host:hostport
    # Specifies that connections to the given TCP port or Unix socket on the
    # remote (server) host are to be forwarded to the local side.
    TUN="${TUN} -R 80:localhost:80"
fi

if bashio::config.true 'tunnel_https'; then
    TUN="${TUN} -R 443:localhost:443"
fi

if ! bashio::config.equals 'socks_port' 0; then
    # -D [bind_address:]port
    # Currently the SOCKS4 and SOCKS5 protocols are supported, and ssh will act
    # as a SOCKS server.
    TUN="${TUN} -D *:$(bashio::config 'socks_port')"
fi

if ! bashio::config.is_empty 'advanced'; then
    TUN="${TUN} $(bashio::config 'advanced')"
fi

SRV="$(bashio::config 'ssh_user')@$(bashio::config 'ssh_host')"

if ! bashio::config.equals 'ssh_port' 22; then
    SRV="-p $(bashio::config 'ssh_port') ${SRV}"
fi

set +e

while true
do
    if ! bashio::config.is_empty 'before'; then
        CMD="${BASE} ${SRV} $(bashio::config 'before')"
        bashio::log.info "[ $(date +'%m-%d-%Y') ] run: ${CMD}"
        eval $CMD
    fi

    CMD="${BASE} ${TUN} ${SRV}"
    bashio::log.info "[ $(date +'%m-%d-%Y') ] run tunnel: ${CMD}"
    eval $CMD

    sleep 30
done
