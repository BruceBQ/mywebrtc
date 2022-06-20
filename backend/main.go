package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/aler9/gortsplib"
	"github.com/aler9/gortsplib/pkg/url"
	"github.com/deepch/vdk/av"
	"github.com/deepch/vdk/codec/h264parser"
	"github.com/deepch/vdk/format/rtspv2"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media"
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

	c := gortsplib.Client{}
	u, err := url.Parse("rtsp://vietbq@centic.vn:centic.vn@14.224.162.205:30554/cameras/618320a640275d57b16c3e94/channels/1")
	if err != nil {
		panic(err)
	}

	if err := c.Start(u.Scheme, u.Host); err != nil {
		panic(err)
	}
	defer c.Close()

	tracks, baseURL, _, err := c.Describe(u)
	if err != nil {
		panic(err)
	}

	// find the H264 track
	h264TrackID, _ := func() (int, *gortsplib.TrackH264) {
		for i, track := range tracks {
			if h264track, ok := track.(*gortsplib.TrackH264); ok {
				return i, h264track
			}
		}
		return -1, nil
	}()
	if h264TrackID < 0 {
		panic("H264 track not found")
	}

	h264channel := make(chan *av.Packet, 100)

	c.OnPacketRTP = func(ctx *gortsplib.ClientOnPacketRTPCtx) {
		if ctx.TrackID != h264TrackID {
			return
		}

		if ctx.H264NALUs == nil {
			return
		}

		// h264channel <- ctx

		// fmt.Println(ctx.Packet)

	}

	c.SetupAndPlay(tracks, baseURL)
	go func() {
		if err := c.Wait(); err == nil {
			fmt.Println("Err RTSP", err)
		}
	}()

	client, err := rtspv2.Dial(rtspv2.RTSPClientOptions{
		URL:              "rtsp://vietbq@centic.vn:centic.vn@14.224.162.205:30554/cameras/618320a640275d57b16c3e94/channels/1",
		DialTimeout:      3 * time.Second,
		ReadWriteTimeout: 3 * time.Second,
		Debug:            false,
	})

	if err != nil {
		panic(err)
	}

	defer client.Close()

	go func() {
		for {
			packetAV := <-client.OutgoingPacketQueue
			h264channel <- packetAV
		}
	}()
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

	// iceConnectedCtx, iceConnectedCtxCancel := context.WithCancel(context.Background())
	_, iceConnectedCtxCancel := context.WithCancel(context.Background())

	videoTrack, err := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeH264}, "video", "vietbq")
	// videoTrack, err := webrtc.NewTrackLocalStaticRTP(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeH264}, "video", "vietbq")
	fmt.Println("videoTrack", videoTrack)
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
		// file, err := os.Open("output.h264")
		// if err != nil {
		// 	panic(err)
		// }

		// h264, err := h264reader.NewReader(file)
		// <-iceConnectedCtx.Done()

		// h264FrameDuration := time.Millisecond * 33

		// ticker := time.NewTicker(h264FrameDuration)
		// for ; true; <-ticker.C {
		// 	nal, err := h264.NextNAL()
		// 	if err == io.EOF {
		// 		fmt.Printf("All video frames parsed and sent")
		// 		os.Exit(0)
		// 	}
		// 	if err != nil {
		// 		panic(err)
		// 	}

		// 	if err = videoTrack.WriteSample(media.Sample{Data: nal.Data, Duration: h264FrameDuration}); err != nil {
		// 		panic(err)
		// 	}
		// }

		for {
			pkt := <-client.OutgoingPacketQueue
			// rtpBuf := make([]byte, 1400)

			// result, err := rtpPacket.Packet.MarshalTo(rtpBuf)
			// fmt.Println("Result", result)
			// if err != nil {
			// 	fmt.Println("Err marshal to", err, result)
			// }
			nalus, _ := h264parser.SplitNALUs(pkt.Data)
			for _, nalu := range nalus {
				naltype := nalu[0] & 0x1f
				// fmt.Println("naltype", naltype)
				// fmt.Println("duration", pkt.Duration)
				if naltype == 5 {
					// codec := client.CodecData.(h264parser.CodecData)
					codec := client.CodecData[0].(h264parser.CodecData)
					err = videoTrack.WriteSample(media.Sample{Data: append([]byte{0, 0, 0, 1}, bytes.Join([][]byte{codec.SPS(), codec.PPS(), nalu}, []byte{0, 0, 0, 1})...), Duration: pkt.Duration})

				} else if naltype == 1 {

					err = videoTrack.WriteSample(media.Sample{Data: append([]byte{0, 0, 0, 1}, nalu...), Duration: pkt.Duration})
				}

				if err != nil {
					fmt.Println(err)
				}
			}
			// if err := videoTrack.WriteSample(media.Sample{Data: pkt.Data, Duration: pkt.Duration}); err != nil {
			// 	fmt.Println("Err Write Sample", err)
			// }
		}
		// <-iceConnectedCtx.Done()

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
