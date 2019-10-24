package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/garyburd/redigo/redis"
	"github.com/gorilla/websocket"
	"log"
	"web-chats/models"
)

var upgrader = websocket.Upgrader{}

type MyWebsocketContro struct {
	beego.Controller
}

func (c *MyWebsocketContro) Get() {
	roomKey := c.GetString(":room_key")
	UUID := c.GetString(":user_key")
	ws, err := upgrader.Upgrade(c.Ctx.ResponseWriter, c.Ctx.Request, nil)
	if err != nil {
		log.Fatal(err)
	}
	// 登记接入的用户
	rConn, err := redis.Dial("tcp", "127.0.0.1:6379")
	if err != nil {
		fmt.Println("redis连接失败：", err)
		return
	}
	defer rConn.Close()
	// 先从 redis 中找到对应房间的数据
	res, _ := redis.String(rConn.Do("HGET", "room_persons", roomKey))
	// 如果找不到，就新建一个数据
	if len(res) == 0 {
		var customerList []string
		customerList = append(customerList, UUID)
		customerListStr, _ := json.Marshal(customerList)
		rConn.Do("HSET", "room_persons", roomKey, customerListStr)
	}else {
		// 如果能找到，就更新其中的数据
		var customerList []string
		err = json.Unmarshal([]byte(res), &customerList)
		var hasCustomer bool
		for _, customerUUID := range customerList {
			// 如果该用户已经进入房间了，就不再重复登记
			if customerUUID == UUID {
				hasCustomer = true
				break
			}
		}
		if hasCustomer == false {
			customerList = append(customerList, UUID)
			customerListStr, _ := json.Marshal(customerList)
			rConn.Do("HSET", "room_persons", roomKey, customerListStr)
		}
	}

	// 连接断开时该做的操作
	defer func() {
		fmt.Print("连接已断开")
		ws.Close()
	}()
	clients[ws] = roomKey
	// 存储当前房间
	all_room[roomKey] = ws

	for {
		var msg models.Message
		// 以 json 格式读取数据并映射到 Message 对象
		err := ws.ReadJSON(&msg)
		msg.RoomKey = roomKey
		if err != nil {
			log.Printf("error: %v", err)
			delete(clients, ws)
			break
		}
		// 将新接收的消息发送到广播频道
		broadcast <- msg
	}
}
