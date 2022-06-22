package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/deepch/vdk/av"
	"github.com/deepch/vdk/codec/h264parser"
	"github.com/deepch/vdk/format/rtspv2"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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

type StreamSdp struct {
	Type string `json:"type"`
	Sdp  string `json:"sdp"`
}

type Stream struct {
}
type Server struct {
	stream map[string]Viewer
}

type Viewer struct {
	c chan av.Packet
}

func main() {
	server := &Server{}

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
	// h264channel := make(chan *av.Packet, 100)
	go func() {
		for {
			packetAV := <-client.OutgoingPacketQueue
			for _, v := range server.stream {
				v.c <- *packetAV
			}
		}
	}()

	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowAllOrigins: true,
		AllowMethods:    []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
	}))

	r.POST("/stream/webrtc", func(c *gin.Context) {
		var stream StreamSdp
		if err := c.BindJSON(&stream); err != nil {
			return
		}

		offer := webrtc.SessionDescription{}
		if stream.Type == "offer" {
			offer.Type = 1
		}

		offer.SDP = stream.Sdp
		pc, err := webrtc.NewPeerConnection(webrtc.Configuration{
			ICEServers: []webrtc.ICEServer{
				{
					URLs: []string{"stun:stun.l.google.com:19302"},
				},
			},
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"detail": err.Error(),
			})
		}

		videoTrack, err := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: webrtc.MimeTypeH264}, "video", "vietbq")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"detail": err.Error(),
			})
		}

		if _, err := pc.AddTrack(videoTrack); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"detail": err.Error(),
			})
		}

		pc.OnICEConnectionStateChange(func(is webrtc.ICEConnectionState) {
			if is == webrtc.ICEConnectionStateDisconnected {
				pc.Close()
			}
		})

		if err := pc.SetRemoteDescription(offer); err != nil {
			panic(err)
		}

		answer, err := pc.CreateAnswer(nil)
		if err != nil {
			panic(err)
		}

		gatherComplete := webrtc.GatheringCompletePromise(pc)

		if err := pc.SetLocalDescription(answer); err != nil {
			panic(err)
		}
		<-gatherComplete
		local := pc.LocalDescription()
		c.JSON(http.StatusOK, gin.H{
			"type": local.Type.String(),
			"sdp":  local.SDP,
		})
		fmt.Println("vietbnq")
		go func() {
			id := uuid.NewString()
			fmt.Println("id", id)
			ch := make(chan av.Packet, 100)
			if server.stream == nil {
				server.stream = map[string]Viewer{}
			}

			server.stream[id] = Viewer{c: ch}
			defer pc.Close()
			for {
				select {
				case pkt := <-ch:
					// fmt.Println(pkt.Idx)
					nalus, _ := h264parser.SplitNALUs(pkt.Data)
					for _, nalu := range nalus {
						naltype := nalu[0] & 0x1f
						if naltype == 5 {
							codec := client.CodecData[0].(h264parser.CodecData)
							err = videoTrack.WriteSample(media.Sample{Data: append([]byte{0, 0, 0, 1}, bytes.Join([][]byte{codec.SPS(), codec.PPS(), nalu}, []byte{0, 0, 0, 1})...), Duration: pkt.Duration})

						} else if naltype == 1 {

							err = videoTrack.WriteSample(media.Sample{Data: append([]byte{0, 0, 0, 1}, nalu...), Duration: pkt.Duration})
						}

						if err != nil {
							fmt.Println(err)
						}
					}
				}
			}
		}()
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
