package main

import (
	"fmt"
	"net/http"

	"github.com/aler9/gortsplib"
	"github.com/aler9/gortsplib/pkg/base"
	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v3"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Camera struct {
	RtspLink string
}

func main() {
	camera := Camera{
		RtspLink: "rtsp://vietbq@centic.vn:centic.vn@doxe.danang.gov.vn:30554/cameras/60c1e33937eb18a230bd0e7f/channels/1",
	}

	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{URLs: []string{"stun:stun.l.google.com:19302"}},
		},
	}

	api := webrtc.NewAPI()
	peerConnection, err := api.NewPeerConnection(config)
	if err != nil {
		fmt.Println("NewPeerConnection Err:", err)
		return
	}

	fmt.Printf("NewPeerConnection: %v\n", peerConnection)

	client := gortsplib.Client{}

	u, err := base.ParseURL(camera.RtspLink)
	if err != nil {
		panic(err)
	}

	err = client.Start(u.Scheme, u.Host)
	if err != nil {
		panic(err)
	}
	defer client.Close()

	// find published tracks
	tracks, baseURL, _, err := client.Describe(u)
	if err != nil {
		panic(err)
	}

	for _, track := range tracks {
		fmt.Println(track.MediaDescription())
		attributes := track.MediaDescription().Attributes
		for _, attribute := range attributes {
			fmt.Println("Attribute:", attribute.Value)
		}
		// fmt.Println("Track:", track.MediaDescription().Attributes)
	}
	offer, err := peerConnection.CreateOffer(nil)
	fmt.Println(offer)

	client.OnPacketRTP = func(ctx *gortsplib.ClientOnPacketRTPCtx) {
		// fmt.Println("trackId:", ctx.Packet)
	}
	// fmt.Println("baseURL", baseURL)
	err = client.SetupAndPlay(tracks, baseURL)
	if err != nil {
		panic(err)
	}
	panic(client.Wait())

}
