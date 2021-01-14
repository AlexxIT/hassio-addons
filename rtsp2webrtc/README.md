# RTSP2WebRTC

[Hass.io](https://www.home-assistant.io/hassio/) addon allows you to watch an [RTSP](https://en.wikipedia.org/wiki/Real_Time_Streaming_Protocol) stream in real time using [WebRTC](https://en.wikipedia.org/wiki/WebRTC) technology.

Based on:
 - [Pion](https://github.com/pion/webrtc) - pure Go implementation of WebRTC 
 - [RTSPtoWebRTC](https://github.com/deepch/RTSPtoWebRTC) - Go app by [@deepch](https://github.com/deepch) and [@vdalex25](https://github.com/vdalex25)
 
Why WebRTC:
- works in any modern browser, even on mobiles
- the only browser technology with minimal camera stream delays (0.5 seconds and below)
- works well with unstable channel
- does not use transcoding and does not load the CPU
- support camera stream with sound

Tested on:
- macOS: Google Chrome, Safari
- Windows: Google Chrome
- Android: Google Chrome

Limitations:
- works only with H.264 camaras
- for external access you need a white IP address (without provider NAT), dynamic IP is also supported

Known work cameras:
- Sonoff GK-200MP2-B (support sound)  
   `rtsp://rtsp:12345678@192.168.1.123:554/av_stream/ch0`  
   `rtsp://rtsp:12345678@192.168.1.123:554/av_stream/ch1`
- EZVIZ C3S  
   `rtsp://admin:111111@192.168.1.123:554/h264/ch01/main/av_stream`  
   `rtsp://admin:111111@192.168.1.123:554/h264/ch01/sub/av_stream`
- Hikvision: DS-2CD2T47G1-L, DS-2CD1321-I, DS-2CD2143G0-IS  
   `rtsp://user:pass@192.168.1.123:554/ISAPI/Streaming/Channels/102`
- Reolink: RLC-410, RLC-410W, E1 Pro, 4505MP
- TP-Link: Tapo C200

Support external camera access. You need to forward UDP ports 50000-50009 to Hass.io server on your router.

Addon Web UI works via Hass.io Ingress technology. Access keys to cameras are transmitted through Hass authorization. If you set up HTTPS and two-factor authentication in Hass, this increases security.

50000-50009 ports are used only during video streaming. At each start of the streaming, a random port is occupied. The port is released when the streaming ends. The data should theoretically be encrypted, but I haven't tested :)