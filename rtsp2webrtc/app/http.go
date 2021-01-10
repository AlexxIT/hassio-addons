package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sort"
	"time"

	"github.com/deepch/vdk/av"

	"github.com/gin-gonic/gin"
)

type JCodec struct {
	Type string
}

func serveHTTP() {
	router := gin.Default()
	router.LoadHTMLGlob("web/templates/*")
	router.GET("/", func(c *gin.Context) {
		fi, all := Config.list()
		suuid := c.DefaultQuery("player", fi)
		sort.Strings(all)
		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"port":     Config.Server.HTTPPort,
			"suuid":    suuid,
			"suuidMap": all,
			"version":  time.Now().String(),
		})
	})
	router.POST("/recive", HTTPAPIServerStreamWebRTC)
	router.GET("/codec/:uuid", func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")

		if Config.ext(c.Param("uuid")) {
			codecs := Config.coGe(c.Param("uuid"))
			if codecs == nil {
				return
			}
			var tmpCodec []JCodec
			for _, codec := range codecs {
				if codec.Type() != av.H264 && codec.Type() != av.PCM_ALAW && codec.Type() != av.PCM_MULAW && codec.Type() != av.OPUS {
					log.Println("Codec Not Supported WebRTC ignore this track", codec.Type())
					continue
				}
				if codec.Type().IsVideo() {
					tmpCodec = append(tmpCodec, JCodec{Type: "video"})
				} else {
					tmpCodec = append(tmpCodec, JCodec{Type: "audio"})
				}
			}
			b, err := json.Marshal(tmpCodec)
			if err == nil {
				_, err = c.Writer.Write(b)
				if err != nil {
					log.Println("Write Codec Info error", err)
					return
				}
			}
		}
	})
	router.StaticFS("/static", http.Dir("web/static"))
	err := router.Run(Config.Server.HTTPPort)
	if err != nil {
		log.Fatalln("Start HTTP Server error", err)
	}
}

//HTTPAPIServerStreamWebRTC stream video over WebRTC
func HTTPAPIServerStreamWebRTC(c *gin.Context) {
	if !Config.ext(c.PostForm("suuid")) {
		log.Println("Stream Not Found")
		return
	}
	codecs := Config.coGe(c.PostForm("suuid"))
	if codecs == nil {
		log.Println("Stream Codec Not Found")
		return
	}
	var AudioOnly bool
	if len(codecs) == 1 && codecs[0].Type().IsAudio() {
		AudioOnly = true
	}
	muxerWebRTC := NewMuxer()
	//muxerWebRTC.pc.api.settingEngine.ephemeralUDP.PortMin = 50000;
	//muxerWebRTC.pc.api.settingEngine.ephemeralUDP.PortMin = 50009;
	answer, err := muxerWebRTC.WriteHeader(codecs, c.PostForm("data"))
	if err != nil {
		log.Println("WriteHeader", err)
		return
	}
	_, err = c.Writer.Write([]byte(answer))
	if err != nil {
		log.Println("Write", err)
		return
	}
	go func() {
		cid, ch := Config.clAd(c.PostForm("suuid"))
		defer Config.clDe(c.PostForm("suuid"), cid)
		defer muxerWebRTC.Close()
		var videoStart bool
		noVideo := time.NewTimer(10 * time.Second)
		for {
			select {
			case <-noVideo.C:
				log.Println("noVideo")
				return
			case pck := <-ch:
				if pck.IsKeyFrame || AudioOnly {
					noVideo.Reset(10 * time.Second)
					videoStart = true
				}
				if !videoStart && !AudioOnly {
					continue
				}
				err = muxerWebRTC.WritePacket(pck)
				if err != nil {
					log.Println("WritePacket", err)
					return
				}
			}
		}
	}()
}
