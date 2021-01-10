package main

import (
	"bytes"
	"encoding/base64"
	"errors"
	"github.com/pion/interceptor"
	"log"
	"time"

	"github.com/pion/webrtc/v3"

	"github.com/deepch/vdk/av"
	"github.com/deepch/vdk/codec/h264parser"
	"github.com/pion/webrtc/v3/pkg/media"
)

var (
	ErrorNotFound          = errors.New("WebRTC Stream Not Found")
	ErrorCodecNotSupported = errors.New("WebRTC Codec Not Supported")
	ErrorClientOffline     = errors.New("WebRTC Client Offline")
	ErrorNotTrackAvailable = errors.New("WebRTC Not Track Available")
	ErrorIgnoreAudioTrack  = errors.New("WebRTC Ignore Audio Track codec not supported WebRTC support only PCM_ALAW or PCM_MULAW")
)

type Muxer struct {
	streams   map[int8]*Stream
	status    webrtc.ICEConnectionState
	stop      bool
	pc        *webrtc.PeerConnection
	ClientACK *time.Timer
	StreamACK *time.Timer
}
type Stream struct {
	codec av.CodecData
	ts    time.Duration
	track *webrtc.TrackLocalStaticSample
}

func NewMuxer() *Muxer {
	tmp := Muxer{ClientACK: time.NewTimer(time.Second * 20), StreamACK: time.NewTimer(time.Second * 20), streams: make(map[int8]*Stream)}
	go tmp.WaitCloser()
	return &tmp
}

func NewPeerConnection(configuration webrtc.Configuration) (*webrtc.PeerConnection, error) {
	m := &webrtc.MediaEngine{}
	if err := m.RegisterDefaultCodecs(); err != nil {
		return nil, err
	}

	i := &interceptor.Registry{}
	if err := webrtc.RegisterDefaultInterceptors(m, i); err != nil {
		return nil, err
	}

	s := webrtc.SettingEngine{}
	s.SetEphemeralUDPPortRange(50000, 50009)
	log.Println("Set UDP ports to 50000..50009")
	api := webrtc.NewAPI(webrtc.WithMediaEngine(m), webrtc.WithInterceptorRegistry(i), webrtc.WithSettingEngine(s))
	return api.NewPeerConnection(configuration)
}

func (element *Muxer) WriteHeader(streams []av.CodecData, sdp64 string) (string, error) {
	var WriteHeaderSuccess bool
	if len(streams) == 0 {
		return "", ErrorNotFound
	}
	sdpB, err := base64.StdEncoding.DecodeString(sdp64)
	if err != nil {
		return "", err
	}
	offer := webrtc.SessionDescription{
		Type: webrtc.SDPTypeOffer,
		SDP:  string(sdpB),
	}

	peerConnection, err := NewPeerConnection(webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	})
	if err != nil {
		return "", err
	}
	defer func() {
		if !WriteHeaderSuccess {
			err = element.Close()
			if err != nil {
				log.Println(err)
			}
		}
	}()
	for i, i2 := range streams {
		var track *webrtc.TrackLocalStaticSample
		if i2.Type().IsVideo() {
			if i2.Type() == av.H264 {
				track, err = webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{
					MimeType: "video/h264",
				}, "pion-rtsp-video", "pion-rtsp-video")
				if err != nil {
					return "", err
				}
				if _, err = peerConnection.AddTrack(track); err != nil {
					return "", err
				}
			}
		} else if i2.Type().IsAudio() {
			AudioCodecString := webrtc.MimeTypePCMA
			switch i2.Type() {
			case av.PCM_ALAW:
				AudioCodecString = webrtc.MimeTypePCMA
			case av.PCM_MULAW:
				AudioCodecString = webrtc.MimeTypePCMU
			case av.OPUS:
				AudioCodecString = webrtc.MimeTypeOpus
			default:
				log.Println(ErrorIgnoreAudioTrack)
				continue
			}
			track, err = webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{
				MimeType:  AudioCodecString,
				Channels:  uint16(i2.(av.AudioCodecData).ChannelLayout().Count()),
				ClockRate: uint32(i2.(av.AudioCodecData).SampleRate()),
			}, "pion-rtsp-audio", "pion-rtsp-audio")
			if err != nil {
				return "", err
			}
			if _, err = peerConnection.AddTrack(track); err != nil {
				return "", err
			}
		}
		element.streams[int8(i)] = &Stream{track: track, codec: i2}
	}
	if len(element.streams) == 0 {
		return "", ErrorNotTrackAvailable
	}
	peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		element.status = connectionState
		if connectionState == webrtc.ICEConnectionStateDisconnected {
			element.Close()
		}
	})
	peerConnection.OnDataChannel(func(d *webrtc.DataChannel) {
		d.OnMessage(func(msg webrtc.DataChannelMessage) {
			element.ClientACK.Reset(5 * time.Second)
		})
	})

	if err = peerConnection.SetRemoteDescription(offer); err != nil {
		return "", err
	}
	gatherCompletePromise := webrtc.GatheringCompletePromise(peerConnection)
	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		return "", err
	}
	if err = peerConnection.SetLocalDescription(answer); err != nil {
		return "", err
	}
	element.pc = peerConnection
	waitT := time.NewTimer(time.Second * 10)
	select {
	case <-waitT.C:
		return "", errors.New("gatherCompletePromise wait")
	case <-gatherCompletePromise:
		//Connected
	}
	resp := peerConnection.LocalDescription()
	WriteHeaderSuccess = true
	return base64.StdEncoding.EncodeToString([]byte(resp.SDP)), nil

}

func (element *Muxer) WritePacket(pkt av.Packet) (err error) {
	//log.Println("WritePacket", pkt.Time, element.stop, webrtc.ICEConnectionStateConnected, pkt.Idx, element.streams[pkt.Idx])
	var WritePacketSuccess bool
	defer func() {
		if !WritePacketSuccess {
			element.Close()
		}
	}()
	if element.stop {
		return ErrorClientOffline
	}
	//Wait client ICEConnectionStateConnected
	if element.status == webrtc.ICEConnectionStateChecking {
		WritePacketSuccess = true
		return nil
	}
	if element.status != webrtc.ICEConnectionStateConnected {
		return nil
	}
	if tmp, ok := element.streams[pkt.Idx]; ok {
		element.StreamACK.Reset(10 * time.Second)
		if tmp.ts == 0 {
			tmp.ts = pkt.Time
		}
		switch tmp.codec.Type() {
		case av.H264:
			codec := tmp.codec.(h264parser.CodecData)
			if pkt.IsKeyFrame {
				pkt.Data = append([]byte{0, 0, 0, 1}, bytes.Join([][]byte{codec.SPS(), codec.PPS(), pkt.Data[4:]}, []byte{0, 0, 0, 1})...)
			} else {
				pkt.Data = pkt.Data[4:]
			}
		case av.PCM_MULAW:
			pkt.Data = pkt.Data[4:]
		case av.PCM_ALAW:
			pkt.Data = pkt.Data[4:]
		case av.OPUS:
			pkt.Data = pkt.Data[4:]
		default:
			return ErrorCodecNotSupported
		}
		err = tmp.track.WriteSample(media.Sample{Data: pkt.Data, Duration: pkt.Time - element.streams[pkt.Idx].ts})
		if err == nil {
			element.streams[pkt.Idx].ts = pkt.Time
			WritePacketSuccess = true
		}
		return err
	} else {
		WritePacketSuccess = true
		return nil
	}
}
func (element *Muxer) WaitCloser() {
	waitT := time.NewTimer(time.Second * 10)
	for {
		select {
		case <-waitT.C:
			if element.stop {
				return
			}
			waitT.Reset(time.Second * 10)
		case <-element.StreamACK.C:
			log.Println("Stream Not Send Video Close")
			element.Close()
		case <-element.ClientACK.C:
			log.Println("Client Not Send ACK (probably the browser is minimized) or tab not active Close client")
			element.Close()
		}
	}
}
func (element *Muxer) Close() error {
	element.stop = true
	if element.pc != nil {
		err := element.pc.Close()
		if err != nil {
			return err
		}
	}
	return nil
}
