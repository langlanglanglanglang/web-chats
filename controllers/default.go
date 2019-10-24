package controllers

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/garyburd/redigo/redis"
	"strings"
	"time"
)

type MainController struct {
	beego.Controller
}

// 创建房间接口
type CreateRoomController struct {
	beego.Controller
}

// 获取房间列表
type GetRoomList struct {
	beego.Controller
}

// 用户注册接口
type Register struct {
	beego.Controller
}

// 获取用户信息接口
type GetUserInfo struct {
	beego.Controller
}

type Room struct {
	RoomName     string
	CreateUUID   string
	FromDate     string
	RoomHashName string
	AllHeadURL   []string
}

type ReturnInfo struct {
	Code string
	Msg  string
}

type User struct {
	Name    string // 用户名
	UUID    string // 用户唯一 ID
	HeadURL string // 用户头像URL
}

func (c *MainController) Get() {
	c.TplName = "index.html"
}

// 创建房间接口
func (this *CreateRoomController) Post() {
	var room Room
	err := json.Unmarshal(this.Ctx.Input.RequestBody, &room)
	if err != nil {
		fmt.Println("json 格式化错误", err)
	}
	room.FromDate = time.Now().Format("2006-01-02 15:04:05")
	// 去除房间名首尾空格
	room.RoomName = strings.Trim(room.RoomName, " ")
	// 根据房间名生成房间号
	roomCode := MD5(room.RoomName)
	room.RoomHashName = roomCode
	// 房间信息转 json
	roomStr, _ := json.Marshal(room)
	// 获取 redis 连接
	rConn, err := redis.Dial("tcp", "127.0.0.1:6379")
	if err != nil {
		fmt.Println("连接 redis 失败：", err)
		return
	}
	defer rConn.Close()

	// 检查该房间是否存在
	res, _ := redis.Bool(rConn.Do("HEXISTS", "room_key", roomCode))
	if res {
		returnInfo := ReturnInfo{"-1", "创建房间失败，该房间名已存在，换一个吧"}
		this.Data["json"] = &returnInfo
		this.ServeJSON()
		return
	}
	// 房间信息存入redis
	// 房间名的 HASH 值，防止房间名重复
	rConn.Do("HSET", "room_key", roomCode, 1)
	// 房间信息 JSON 字符串
	rConn.Do("LPUSH", "room_list", roomStr)
	returnInfo := ReturnInfo{"0", "创建房间成功"}
	this.Data["json"] = &returnInfo
	this.ServeJSON()
}

func MD5(s string) string {
	m := md5.New()
	m.Write([]byte(s))
	return hex.EncodeToString(m.Sum(nil))
}

// 获取房间列表
func (this *GetRoomList) Get() {
	rConn, err := redis.Dial("tcp", "127.0.0.1:6379")
	if err != nil {
		fmt.Println("连接 redis 失败：", err)
		return
	}
	// 获取 redis 中所有房间
	roomListStr, err := redis.Values(rConn.Do("LRANGE", "room_list", 0, -1))
	if err != nil {
		fmt.Println("查询失败：", err)
		return
	}
	var roomList []Room
	// 遍历解析 json ，获取房间信息
	for _, roomStr := range roomListStr {
		var tmpRoom Room
		err := json.Unmarshal([]byte(roomStr.([]uint8)), &tmpRoom)
		if err != nil {
			fmt.Println("JSON 格式化失败", err)
			continue
		}
		roomList = append(roomList, tmpRoom)
	}
	// 获取所有房间内当前在线用户信息
	// FIXME 功能尚未完成，还需查询用户头像，组装数据

	if err != nil {
		fmt.Println("查询失败：", err)
		return
	}
	// 获取房间内在线的用户，将他们的头像组装到数据中
	for index, value := range roomList {
		var customerList []string
		roomPersons, _ := redis.String(rConn.Do("HGET", "room_persons", value.RoomHashName))
		json.Unmarshal([]byte(roomPersons), &customerList)
		roomList[index].AllHeadURL = make([]string, 9)
		for n, customerCode := range customerList {
			if n <= 9 {
				res, _ := redis.String(rConn.Do("HGET", "user_list", customerCode))
				var customer User
				json.Unmarshal([]byte(res), &customer)
				roomList[index].AllHeadURL = append(roomList[index].AllHeadURL, customer.HeadURL)
			}
		}
	}
	this.Data["json"] = &roomList
	this.ServeJSON()
}

// 用户初次进入注册
func (this *Register) Post() {
	var user User
	// 接收参数
	err := json.Unmarshal(this.Ctx.Input.RequestBody, &user)
	if err != nil {
		fmt.Println("Register：json 格式化错误", err)
	}
	rConn, err := redis.Dial("tcp", "127.0.0.1:6379")
	if err != nil {
		fmt.Println("连接 redis 失败：", err)
		return
	}
	// 去除昵称首尾空格
	user.Name = strings.Trim(user.Name, " ")
	fmt.Println("用户ID:", user.UUID)
	res, _ := redis.Bool(rConn.Do("HEXISTS", "user_list", user.UUID))
	if res {
		returnInfo := ReturnInfo{"-1", "注册失败，该昵称已存在，换一个吧"}
		this.Data["json"] = &returnInfo
		this.ServeJSON()
		return
	}
	user.HeadURL = "/static/img/default.jpeg" // 默认头像
	userStr, _ := json.Marshal(user)
	rConn.Do("HSET", "user_list", user.UUID, userStr)
	returnInfo := ReturnInfo{"0", "注册成功"}
	this.Data["json"] = &returnInfo
	this.ServeJSON()
}

func (this *GetUserInfo) Get() {
	userKey := this.GetString("user_key")
	var user User
	rConn, err := redis.Dial("tcp", "127.0.0.1:6379")
	if err != nil {
		fmt.Println("连接redis失败：", err)
		return
	}
	fmt.Println("userKey:", userKey)
	userStr, _ := redis.String(rConn.Do("HGET", "user_list", userKey))
	json.Unmarshal([]byte(userStr), &user)
	this.Data["json"] = &user
	this.ServeJSON()
}