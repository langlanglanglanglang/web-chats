new Vue({
    el: '#app',
    delimiters: ['${', '}'],
    data: {
        show_chat: false,
        full_height: document.documentElement.clientHeight,
        showCreateRoom: false,
        showRegister: false,
        nick_name: '',
        room_list: null,
        name: '',
        head_url: '',
        // 当前在聊的房间
        this_room_key: '',
        // 当前输入框内容
        this_message: '',
        // 当前聊天内容列表
        this_message_list: [],
        // 所有房间的聊天记录
        all_message_list: {},
        // 所有 ws 连接
        all_ws: {},
        form: {
            name: ''
        }
    },
    methods: {
        //创建房间
        create_room:function () {
            var self = this;
            axios.post("/create_room", JSON.stringify({
                RoomName: self.form.name,
                CreateUUID: self.uuid
            })).then(function (res) {
                console.log(res);
                self.$message({
                    message: res.data.Msg,
                    type: res.data.Code == 0 ? 'success' : 'error'
                });
                self.showCreateRoom = false;
                // 成功创建后清空房间名
                if (res.data.Code == 0){
                    self.form.name = "";
                    self.get_room_list();
                }
            })
        },
        get_uuid:function () {
            var self = this;
            var uuid;
            if (!$cookies.isKey("UUID")) {
                self.showRegister = true;
            }else{
                uuid = $cookies.get("UUID");
                self.name = $cookies.get("Name");
            }
            return uuid
        },
        S4:function() {
            return (((1+Math.random())*0x10000)|0).toString(16).substring(1);
        },
        guid:function() {
            return (this.S4() + this.S4() + "-" + this.S4() + "-" + this.S4() + "-" + this.S4() + "-" + this.S4() + this.S4() + this.S4());
        },
        get_room_list:function(){
            var self = this;
            axios.get("/get_room_list").then(function (res) {
                self.room_list = res.data
            })
        },
        // 初次使用注册
        register:function(){
            var uuid;
            var self = this;
            uuid = self.guid();
            if (self.nick_name == ''){
                self.$message({
                    message: "昵称不可为空",
                    type: "warning"
                });
                return
            }
            self.uuid = uuid;
            self.name = self.nick_name;
            axios.post("/register", JSON.stringify({
                Name: self.nick_name,
                UUID: self.uuid
            })).then(function (res) {
                console.log(res);
                self.$message({
                    message: res.data.Msg,
                    type: res.data.Code == 0 ? 'success' : 'error'
                });
                if (res.data.Code == 0){
                    self.showRegister = false;
                    $cookies.set("UUID", uuid, 365 * 24 * 3600);
                    $cookies.set("Name", self.nick_name, 365 * 24 * 3600);
                }else {
                    self.uuid = null
                    self.name = null
                }
            })
        },
        open_room:function (room_id) {
            const self = this;
            if (!this.all_ws.hasOwnProperty(room_id)){
                let ws = new WebSocket('ws://' + window.location.host + '/ws/' + room_id + '/' + this.uuid)
                this.all_message_list[room_id] = []
                ws.onmessage = function (msg) {
                    let data = JSON.parse(msg.data);
                    if (data.username == self.name){
                        data["who"] = 'me'
                    }else{
                        data["who"] = 'he'
                    }
                    if (self.all_message_list.hasOwnProperty(data.room_key)){
                        self.all_message_list[data.room_key].push(data)
                    }else {
                        self.all_message_list[data.room_key] = []
                        self.all_message_list[data.room_key].push(data)
                    }
                    self.$nextTick(function (e) {
                        self.$refs.message_list.scrollTop = self.$refs.message_list.scrollHeight;
                    })
                }
                this.all_ws[room_id] = ws
            }
            this.show_chat = true;
            this.this_room_key = room_id
            this.this_message_list = this.all_message_list[room_id]
            this.get_room_list()
        },
        send:function (e) {
            let msg = this.this_message.replace("\n", "");
            if (msg == ""){
                this.this_message = ''
                return
            }
            //组装消息
            let msg1 = {
                username: this.name,
                message: msg,
                room_key: this.this_room_key,
                head_url: this.head_url
            }
            // 发送消息
            this.all_ws[this.this_room_key].send(JSON.stringify(msg1))
            this.this_message = ''
        },
        get_user_info(){
            var self = this;
            axios.get("/get_user_info", {params:{user_key: self.uuid}}).then(function (res) {
                self.head_url = res.data.HeadURL
            })
        }
    },
    mounted: function(){
        this.uuid = this.get_uuid();
        this.get_room_list()
        this.get_user_info()
    },
})
