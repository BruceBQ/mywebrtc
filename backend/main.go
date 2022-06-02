package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func main() {
	r := gin.Default()

	r.GET("/ws", func(c *gin.Context) {
		ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			fmt.Println("upgrader err:", err)
		}
		defer ws.Close()
		for {
			mt, message, err := ws.ReadMessage()
			if err != nil {
				fmt.Println(err)
				break
			}
			fmt.Println(string(message))
			if string(message) == "ping" {
				message = []byte("pong")
			}
			err = ws.WriteMessage(mt, message)
			if err != nil {
				fmt.Println(err)
				break
			}
		}
	})

	r.Run(":3001")
}
