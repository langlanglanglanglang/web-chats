package routers

import (
	"github.com/astaxie/beego"
	"web-chats/controllers"
)

func init() {
	beego.Router("/", &controllers.MainController{})
	beego.Router("/ws/:room_key/:user_key", &controllers.MyWebsocketContro{})
	beego.Router("/create_room", &controllers.CreateRoomController{})
	beego.Router("/get_room_list", &controllers.GetRoomList{})
	beego.Router("/register", &controllers.Register{})
	beego.Router("/get_user_info", &controllers.GetUserInfo{})

}
