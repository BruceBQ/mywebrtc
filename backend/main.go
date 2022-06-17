package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media"
	"github.com/pion/webrtc/v3/pkg/media/h264reader"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Camera struct {
	RtspLink string
}

type Stream struct {
	Type string `json:"type"`
	Sdp  string `json:"sdp"`
}

func main() {

	peerConnection, err := webrtc.NewPeerConnection(webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		if err := peerConnection.Close(); err != nil {
			fmt.Printf("cannot close peerConnection: %v\n", err)
		}
	}()

	iceConnectedCtx, iceConnectedCtxCancel := context.WithCancel(context.Background())

	videoTrack, err := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeH264}, "video", "pion")
	if err != nil {
		panic(err)
	}

	rtpSender, err := peerConnection.AddTrack(videoTrack)
	if err != nil {
		panic(err)
	}

	go func() {
		rtcpBuf := make([]byte, 1500)
		for {
			if _, _, err := rtpSender.Read(rtcpBuf); err != nil {
				return
			}
		}
	}()

	go func() {
		file, err := os.Open("output.h264")
		if err != nil {
			panic(err)
		}

		h264, err := h264reader.NewReader(file)
		<-iceConnectedCtx.Done()

		h264FrameDuration := time.Millisecond * 33

		ticker := time.NewTicker(h264FrameDuration)
		for ; true; <-ticker.C {
			nal, err := h264.NextNAL()
			if err == io.EOF {
				fmt.Printf("All video frames parsed and sent")
				os.Exit(0)
			}
			if err != nil {
				panic(err)
			}

			if err = videoTrack.WriteSample(media.Sample{Data: nal.Data, Duration: h264FrameDuration}); err != nil {
				panic(err)
			}
		}
	}()

	peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		fmt.Printf("Connection State has changed %s\n", connectionState.String())
		if connectionState == webrtc.ICEConnectionStateConnected {
			iceConnectedCtxCancel()
		}
	})

	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowAllOrigins: true,
		AllowMethods:    []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
	}))

	gortsplib.
		r.POST("/stream/webrtc", func(c *gin.Context) {
		var stream Stream
		if err := c.BindJSON(&stream); err != nil {
			return
		}

		offer := webrtc.SessionDescription{}
		if stream.Type == "offer" {
			offer.Type = 1
		}

		offer.SDP = stream.Sdp

		if err := peerConnection.SetRemoteDescription(offer); err != nil {
			panic(err)
		}

		answer, err := peerConnection.CreateAnswer(nil)
		if err != nil {
			panic(err)
		}

		gatherComplete := webrtc.GatheringCompletePromise(peerConnection)

		if err := peerConnection.SetLocalDescription(answer); err != nil {
			panic(err)
		}
		<-gatherComplete
		local := peerConnection.LocalDescription()
		c.JSON(http.StatusOK, gin.H{
			"type": local.Type.String(),
			"sdp":  local.SDP,
		})
	})

	r.Run(":3001")
}

type Response struct {
	Id     uint64                     `json:"id"`
	Method string                     `json:"method"`
	Params *webrtc.SessionDescription `json:"params"`
	Result *webrtc.SessionDescription `json:"result"`
}

func readMessage(connection *websocket.Conn, done chan struct{}) {
	defer close(done)

	for {
		_, message, err := connection.ReadMessage()
		if err != nil || err == io.EOF {
			log.Fatal("Error reading: ", err)
		}

		fmt.Printf("recv: %s \n", message)

	}

}
