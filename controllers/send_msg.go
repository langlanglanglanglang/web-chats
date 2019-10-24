package controllers

import (
	"github.com/gorilla/websocket"
	"log"
	"web-chats/models"
)

var (
	clients   = make(map[*websocket.Conn]string)
	all_room  = make(map[string]*websocket.Conn)
	broadcast = make(chan models.Message, 5)
)

func init() {
	go handleMessages()
}

func handleMessages() {
	for {
		// 从广播频道中抓取下一条消息
		msg := <-broadcast
		// 将其发送给当前房间对应的客户端
		for client, room_key := range clients {
			if room_key == msg.RoomKey {
				err := client.WriteJSON(msg)
				if err != nil {
					log.Printf("error: %v", err)
					client.Close()
					delete(clients, client)
				}
			}
		}
	}
}
