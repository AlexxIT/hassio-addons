# RTSP2WebRTC

[Hass.io](https://www.home-assistant.io/hassio/) addon allows you to watch an [RTSP](https://en.wikipedia.org/wiki/Real_Time_Streaming_Protocol) stream in real time using [WebRTC](https://en.wikipedia.org/wiki/WebRTC) technology.

Based on:
 - [Pion](https://github.com/pion/webrtc) - pure Go implementation of WebRTC 
 - [RTSPtoWebRTC](https://github.com/deepch/RTSPtoWebRTC) - Go app by [@deepch](https://github.com/deepch) and [@vdalex25](https://github.com/vdalex25)

Support external camera access. You need to forward UDP ports 50000-50009 to Hass.io server on your router.

Addon Web UI works via Hass.io Ingress technology. Access keys to cameras are transmitted through Hass authorization. If you set up HTTPS and two-factor authentication in Hass, this increases security.

50000-50009 ports are used only during video streaming. At each start of the streaming, a random port is occupied. The port is released when the streaming ends. The data should theoretically be encrypted, but I haven't tested :)