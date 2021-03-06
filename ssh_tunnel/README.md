# SSH Tunnel

Аддон Hass.io для настройки внешнего доступа к **Home Assistant** через внешний сервер по протоколу SSH. Полезно если у вас "серый" IP-адрес.

Для работы у вас уже должен быть арендован сервер с SSH доступом.

По умолчанию аддон прокидывает порты 80 и 443 сервера Hass.io на внешний сервер. А также создаёт SOCKS-прокси на сервере Hass.io.

При первой настройке достаточно настроить `host`, `port` и `user` вашего сервера. 

При запуске в логах выведется публичный SSH-ключ (довольно длинный), например:

```
ssh-rsa AAAAB3NzaC1yc2EA...XfsAODObXZRVMI03 root@2fc7ce79c3de
```

Ключ генерируется случайным образом при каждой установке аддона. Его необходимо **скопировать на ваш публичный сервер** в домашнию директорию пользователя от имени которого вы подключаетесь к северу по SSH:

`~/.ssh/authorized_keys`

Например `/root/.ssh/authorized_keys` или `/home/USERNAME/.ssh/authorized_keys`

При необходимости создать директорию `.ssh` и файл `authorized_keys` - их может не быть.

**Внимание!** Директория `.ssh` считается скрытой.

После обрыва - соединение будет восстанавливаться автоматически. Но прочтите раздел **Восстановление после обрыва**.

## Настройки

- `ssh_host` - адрес публичного сервера
- `ssh_user` - пользователь публичного севрера
- `ssh_port` - порт публичного сервера
- `tunnel_http` - перенапралять HTTP-порт
- `tunnel_https` - перенаправлять HTTPS-порт
- `socks_port` - включить режим SOCKS-прокси на указанном порту сервера Hass.io
- `advanced` - дополнительные параметры ssh-команды
- `before` - см. раздел **Восстановление после обрыва**

**Внимание!** Для прокидывания "привилегированных" портов (например 80 и 443) пользователь публичного сервера должен быть админом. Если у вас с этим проблемы - можете загуглить: **ssh forward privileged ports**.

Пример конфига:

- прокидывает порты 80 и 443 с сервера Hass.io на внешний сервер
- создаёт SOCKS-прокси на сервере Hass.io на порту 1080

```yaml
ssh_host: 87.250.250.242
ssh_user: root
ssh_port: 22
tunnel_http: true
tunnel_https: true
socks_port: 1080
```

## Применение

### Внешний доступ к Home Assistant

Удобно использовать в связке с [Caddy Proxy](https://github.com/bestlibre/hassio-addons/tree/master/caddy_proxy) addon for hass.io. Это легковестный web сервер, который автоматически генерирует сертификаты HTTPS.

Пример конфига Caddy Proxy:

```yaml
homeassistant: sshtunnel.duckdns.org
vhosts": []
raw_config": []
email": sshtunnel@gmail.com
```

E-mail используется при генерации сертификатов. Доменное имя (DNS) необходимо направить на IP-адрес вашего публичного сервера.

**Настройка:**

1. Арендуете VDS сервер
2. Настраиваете аддон SSH Tunnel
3. Создаёте и настраиваете аккаунт Duck DNS
4. Настраиваете аддон Caddy Proxy
5. Настраиваете двухфакторную авторизацию к Home Assistant
6. Пользуетесь безопасным и стабильным внешним доступом к HA

### Прокси для бота Telegram

```yaml
telegram_bot:
- platform: broadcast
  api_key: TELEGRAM_BOT_API_KEY
  allowed_chat_ids: [123456789]
  proxy_url: socks5://172.17.0.1:1080
```

PS: `172.17.0.1` - для стандартной установки hass.io этот IP менять не нужно

### Туннель для любых локальных ресурсов

Например внешний доступ к домашнему OpenVPN серверу, запущенному на роутере.

```yaml
advanced: -R 1194:192.168.1.1:1194
```

PS: Теперь можете подключаться к себе домой по адресу `sshtunnel.duckdns.org:1194`

### Восстановление после обрыва

В аддоне настроена проверка соединения каждые 30 секунд. При 3х неуспешных проверках соединение обрывается.

**Внимание!** При настройках по умолчанию, после обрыва соединения порты на вашем сервере остануться занятыми!

Варианта два:

1. Настроить проверку соединения с клиентом на сервере: [ClientAliveInterval](https://sys-adm.in/os/nix/429-centos-increase-ssh-session-timeout.html)

2. Освобождать порт в начале каждого подключения клиентом

Для этого я сделал опцию `before: fuser -k 443/tcp`. **fuser** - один из способов освободить занятый порт в Ubuntu, **не установлена по умолчанию!**

Лично я использую второй вариант.