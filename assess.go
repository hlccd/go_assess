package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	Database "hlccd"
	"log"
	"net"
	"strconv"
	"time"
)

/*
	监听
	结构各变量描述见readme
*/
type monitor struct {
	conn net.Conn
	id   int64
	sid  int64
	rid  int64
}

//监听表,用于存储每一个仍在监听的客户端链接
var Monitors []monitor = make([]monitor, 0)

/*
	消息结构
	用于接受客户端传来的消息并存放
	结构内各字段描述见readme
*/
type message struct {
	Mes string `json:"mes"`
	Id  int64  `json:"id"`
	Sid int64  `json:"sid"`
	Rid int64  `json:"rid"`
}

/*
	session结构体
	用于验证该账号是否处于登陆状态中
	结构内各字段描述见readme
*/
type Session struct {
	Sid       int64
	Id        int64
	Rid       int64
	State     int
	Timestamp int64
}

//用于存储所有登陆中的账号的session
var SessionS []Session = make([]Session, 0)

//session的ID编号,自1号开始,每有一个登陆则自增一次
var SIDS int64 = int64(1)

/*
	创建Session
	当session存在过程中说明该账号处于登陆状态
	@param id int64 用户id号
	@return sid int64 session指定id,为0则说明创建session失败
*/
func SessionCreate(id int64) (sid int64) {
	sess := Session{
		Sid:       SIDS,
		Id:        id,
		Rid:       0,
		State:     0,
		Timestamp: 0,
	}
	for x := range SessionS {
		if SessionS[x].Id == id {
			/*
				查找SessionS总表中是否存在该账号的session
				如果有说明已经有过登陆,故再次登陆失败
			*/
			return 0
		}
	}
	//之前没有和该id绑定的session,所以可以登陆
	SessionS = append(SessionS, sess)
	SIDS++
	return SIDS - 1
}

/*
	检查sid和id所对应状态的session是否存在
	如果state为2975则不检查状态(该值是临时添加的,所以有一点小问题)
*/
func SessionCheck(sid int64, id int64, state int) (B bool) {
	if state == 2975 {
		for x := range SessionS {
			if SessionS[x].Sid == sid && SessionS[x].Id == id {
				return true
			}
		}
	} else {
		for x := range SessionS {
			if SessionS[x].Sid == sid && SessionS[x].Id == id && SessionS[x].State == state {
				return true
			}
		}
	}

	return false
}

/*
	更改session当前状态
	状态0,仅处于登陆中
	状态1,处于房间中未准备
	状态2,处于房间中已准备
	@param sid int64 session的ID
	@param id int64 用户id
	@param state int 新状态
	@return B bool true表示更改成功,false表示更改失败
*/
func SessionChange(id int64, state int) (B bool) {
	for x := range SessionS {
		if SessionS[x].Id == id {
			SessionS[x].State = state
			return true
		}
	}
	return false
}

/*
	销毁Session
	当session销毁后说明该账号已登出
	可根据时间戳拓展间隔时间的自动清除长时间内无操作的session(此为拓展项,暂时未做)
	@param sid int64 session的ID
	@param id int64 用户id
	@return B bool B为true则表示销毁session成功,为false则表示销毁失败
*/
func SessionDelete(sid int64, id int64) (B bool) {
	for x := range SessionS {
		if SessionS[x].Sid == sid && SessionS[x].Id == id {
			SessionS = append(SessionS[:x], SessionS[x+1:]...)
			return true
		}
	}
	return false
}

/*
	room结构体
	用于储存房间内信息
	结构内各字段描述见readme
*/
type Room struct {
	Room_id  int64 `json:"room_id"`
	Player1  int64 `json:"player1"`
	Player2  int64 `json:"player2"`
	Prepare1 int   `json:"prepare1"`
	Prepare2 int   `json:"prepare2"`
}

//房间总表,用于存储已经建立的房间,随机匹配房间也在其中
var RoomS []Room = make([]Room, 0)

//正常创建房间的序列号,自1起,最高为10000
var RIDS int64 = int64(1)

//低等级匹配的房间的序列号,自10001起,最高为20000
var LowIDS int64 = int64(10001)

//中等级匹配的房间的序列号,自20001起,最高为30000
var MidIDS int64 = int64(20001)

//高等级匹配的房间的序列号,自30001起,最高为40000
var HigIDS int64 = int64(30001)

/*
	随机匹配
	根据胜率寻找空房间
	如果同胜率类型中无空房间,则自行创建一个房间
	找到房间后进入该房间
	@param id int64 用户id
	@param rate float32 用户胜率,用于区别胜率级别
	@return rid int64 房间号,为0说明匹配失败
*/
func Random(id int64, rate float32) (rid int64) {
	if rate <= 0.33 {
		//低胜率房间
		for x := range RoomS {
			if RoomS[x].Player2 == 0 && RoomS[x].Room_id > 10000 && RoomS[x].Room_id <= 20000 {
				rid = RoomS[x].Room_id
				RoomS[x].Player2 = id
				break
			}
		}
	} else if rate <= 0.66 {
		//中胜率房间
		for x := range RoomS {
			if RoomS[x].Player2 == 0 && RoomS[x].Room_id > 20000 && RoomS[x].Room_id <= 30000 {
				rid = RoomS[x].Room_id
				RoomS[x].Player2 = id
				break
			}
		}
	} else {
		//高胜率房间
		for x := range RoomS {
			if RoomS[x].Player2 == 0 && RoomS[x].Room_id > 30000 && RoomS[x].Room_id <= 40000 {
				rid = RoomS[x].Room_id
				RoomS[x].Player2 = id
				break
			}
		}
	}
	if rid == 0 {
		/*
			当房间号为0
			说明之前未能找到仍处于空的同胜率等级的房间
			因此需要自行创建一个同胜率等级的房间并进入
			同时对应胜率房间编号自增一次
		*/
		room := Room{
			Room_id:  0,
			Player1:  0,
			Player2:  0,
			Prepare1: 0,
			Prepare2: 0,
		}
		if rate <= 0.33 {
			rid = LowIDS
			room.Room_id = LowIDS
			room.Player1 = id
			LowIDS++
			RoomS = append(RoomS, room)
		} else if rate <= 0.66 {
			rid = MidIDS
			room.Room_id = MidIDS
			room.Player1 = id
			MidIDS++
			RoomS = append(RoomS, room)
		} else {
			rid = HigIDS
			room.Room_id = HigIDS
			room.Player1 = id
			HigIDS++
			RoomS = append(RoomS, room)
		}
	}
	//如果进入了房间,则更改session的当前状态
	if rid != 0 {
		SessionChange(id, 1)
	}
	return rid
}

/*
	创建Room
	房间的一号玩家为创建者
	一号玩家固定为房主
	@param id int64 用户id号
	@return rid int64 room指定id,为0则说明创建room失败
*/
func RoomCreate(id int64) (rid int64) {
	room := Room{
		Room_id:  RIDS,
		Player1:  id,
		Player2:  0,
		Prepare1: 0,
		Prepare2: 0,
	}
	if SessionChange(id, 1) {
		RoomS = append(RoomS, room)
		RIDS++
		return RIDS - 1
	}
	return 0
}

/*
	进入room
	当二号玩家不存在时,进入成功
	当存在二号玩家时,进入失败
	@param sid int64 session的ID编号
	@param id int64 进入该房间的用户id号
	@return B bool 为true表示进入成功,为false表示进入失败
*/
func RoomEntrance(rid int64, id int64) (B bool) {
	for x := range RoomS {
		if RoomS[x].Room_id == rid {
			if RoomS[x].Player2 == 0 {
				SessionChange(id, 1)
				RoomS[x].Player2 = id
				return true
			} else {
				return false
			}
		}
	}
	return false
}

/*
	进入观战
	当一号玩家存在时才可进入观战
	@param sid int64 session的ID编号
	@param id int64 进入观战用户的id号
	@return B bool 为true表示进入成功,为false表示进入失败
*/
func RoomWatch(rid int64, id int64) (B bool) {
	for x := range RoomS {
		if RoomS[x].Room_id == rid {
			if RoomS[x].Player1 != 0 {
				SessionChange(id, 2975)
				return true
			} else {
				return false
			}
		}
	}
	return false
}

/*
	退出room
	当退出着为一号玩家即房主时,二号玩家自动变成一号玩家
	当二号玩家退出时,房主不变
	当一号玩家退出且无二号玩家时,房间销毁
	@param sid int64 session的ID编号
	@param id int64 退出该房间的用户id号
	@return B bool 为true则说明退出成功,为false则说明退出失败
*/
func RoomQuit(rid int64, id int64) (B bool) {
	for x:=range SessionS{
		if SessionS[x].State==2975 {
			SessionS[x].State=0
			return true
		}
	}
	for x := range RoomS {
		if RoomS[x].Room_id == rid {
			if RoomS[x].Player2 == id {
				RoomS[x].Player2 = 0
				RoomS[x].Prepare2 = 0
				SessionChange(id, 0)
				return true
			} else if RoomS[x].Player1 == id {
				if RoomS[x].Player2 != 0 {
					RoomS[x].Player1 = RoomS[x].Player2
					RoomS[x].Prepare1 = RoomS[x].Prepare2
					RoomS[x].Player2 = 0
					RoomS[x].Prepare2 = 0
					SessionChange(id, 0)
					return true
				} else {
					RoomS = append(RoomS[:x], RoomS[x+1:]...)
					return true
				}
			}
		}
	}
	return false
}

/*
	记录结构
	用于存储从数据库中读取到的记录数据
	结构各字段描述见readme
*/
type Record struct {
	Opponent  int64 `json:"opponent"`
	Win       int   `json:"win"`
	MyGesture int   `json:"my_gesture"`
	OpGesture int   `json:"op_gesture"`
	Timestamp int64 `json:"timestamp"`
}

/*
	创建用户数据总表和打开数据库
	初始时仅创建用户总表
	后续各个账号的历史对局记录表同注册账号时一同创建
*/
func CreateList() (Database.DB, []Database.Table, *sql.DB) {
	//数据库登陆结构信息
	D := Database.DB{
		DriverName: "mysql",
		User:       "root",
		Password:   "2975hLcCd",
		Tcp:        "localhost:3306",
		Name:       "game",
	}
	Tab := make([]Database.Table, 0)
	//0号数据表：用户总列表
	Tab = append(Tab,
		Database.Table{
			Name: "account_list",
			Value: []string{
				"id bigint primary key auto_increment",
				"password varchar(30)",
				"rate float",
				"win bigint",
				"total bigint",
			},
			Annotation: "auto_increment=99999",
		})
	//启动数据库
	db, b, _ := Database.OpenDatabase(D)
	if b {
		//启动成功后创建所有Tab中的数据表
		for x := range Tab {
			Database.CreateTable(db, Tab[x])
		}
	}
	return D, Tab, db
}

/*
	账号系统,其中包括:
	注册
	登陆
	登出
	查看个人历史记录
	relativePath以/account作为开头
*/
func AccountSystem(r *gin.Engine, db *sql.DB) {
	AccountRegister(r, db)
	AccountLogin(r, db)
	AccountLogout(r)
	AccountRecord(r, db)
}

/*
	账号注册
	请求类型:POST
	返回201则表示注册成功
	返回202则表示注册失败
	后续relativePath为"/register"
	Params中仅获取password字段作为密码新账号密码
	新账号ID为数据库内部自增最新值
*/
func AccountRegister(r *gin.Engine, db *sql.DB) {
	r.POST("/account/register", func(c *gin.Context) {
		passwordS := c.Query("password")
		id := int64(0)
		rate := float32(0)
		//添加该账号到数据库中
		b, err := Database.InsertData(db, "account_list", "0,'"+passwordS+"',0.0,0,0")
		if b {
			//获取该账号的id号
			rows, _, _ := Database.SelectLastData(db, "id", "account_list")
			defer rows.Close()
			if rows.Next() {
				err := rows.Scan(&id)
				if err != nil {
					c.JSON(202, gin.H{
						"code":     202,
						"id":       id,
						"password": passwordS,
						"rate":     rate,
					})
					log.Fatal(err)
				} else {
					//未该账号创建对应的历史记录信息表
					Database.CreateTable(db, Database.Table{
						Name: "record" + strconv.FormatInt(id, 10),
						Value: []string{
							"opponent_id bigint",
							"win int",
							"my_gesture int",
							"op_gesture int",
							"timestamp bigint",
						},
						Annotation: "",
					})
					//注册成功,返回该账号的信息
					c.JSON(201, gin.H{
						"code":     201,
						"id":       id,
						"password": passwordS,
						"rate":     rate,
					})
				}
			}
		} else {
			c.JSON(202, gin.H{
				"code":     202,
				"id":       id,
				"password": passwordS,
				"rate":     rate,
			})
			log.Fatal(err)
		}
	})
}

/*
	账号登陆
	请求类型:POST
	返回201则表示登陆成功,创建session成功
	返回202则表示账号密码正确,但未完成session创建
	返回200则表示登陆失败,账号密码错误
	后续relativePath为"/login"
	Params中获取id和password字段作为认证的账号和密码
*/
func AccountLogin(r *gin.Engine, db *sql.DB) {
	r.POST("/account/login", func(c *gin.Context) {
		idS := c.Query("id")
		id, _ := strconv.ParseInt(idS, 10, 64)
		passwordS := c.Query("password")
		sid := int64(0)
		//验证该账号的id和密码是否匹配
		if passwordS == Database.SelectKeyGetFieldS(db, "account_list", "password", "id="+idS) {
			//为这个账号创建一个session作为该账号处于登陆状态的证明
			sid = SessionCreate(id)
			if sid != 0 {
				c.JSON(201, gin.H{
					"code": 201,
					"id":   id,
					"sid":  sid,
				})
			} else {
				c.JSON(202, gin.H{
					"code": 201,
					"id":   id,
					"sid":  sid,
				})
			}
		} else {
			c.JSON(200, gin.H{
				"code": 200,
				"id":   id,
				"sid":  sid,
			})
		}
	})
}

/*
	账号登出
	请求类型:POST
	返回201则表示登出成功,销毁session成功
	返回202则表示登出失败,session不存在,销毁session失败
	后续relativePath为"/logout"
	Params中获取id和sid字段作为登出的session的ID号和用户ID号
*/
func AccountLogout(r *gin.Engine) {
	r.POST("/account/logout", func(c *gin.Context) {
		idS := c.Query("id")
		id, _ := strconv.ParseInt(idS, 10, 64)
		sidS := c.Query("sid")
		sid, _ := strconv.ParseInt(sidS, 10, 64)
		//删除该id和sid对应的session
		B := SessionDelete(sid, id)
		if B {
			c.JSON(201, gin.H{
				"code": 201,
				"id":   id,
				"sid":  sid,
			})
		} else {
			c.JSON(202, gin.H{
				"code": 202,
				"id":   id,
				"sid":  sid,
			})
		}
	})
}

/*
	查看个人历史记录
	请求类型:GET
	返回201则表示查看成功
	返回202则表示认证失败
	后续relativePath为"/record"
	Params中获取id和sid字段用以判断账号是否处于登陆状态中
*/
func AccountRecord(r *gin.Engine, db *sql.DB) {
	r.GET("/account/record", func(c *gin.Context) {
		sidS := c.Query("sid")
		sid, _ := strconv.ParseInt(sidS, 10, 64)
		idS := c.Query("id")
		id, _ := strconv.ParseInt(idS, 10, 64)
		//个人历史记录总表
		records := make([]Record, 0)
		if SessionCheck(sid, id, 2975) {
			//从数据库中查询个人的历史对局记录
			rows, _, _ := Database.SelectAllData(db, "record"+idS, "opponent_id,win,my_gesture,op_gesture,timestamp", "")
			defer rows.Close()
			for rows.Next() {
				record := Record{
					Opponent:  0,
					Win:       0,
					MyGesture: 0,
					OpGesture: 0,
					Timestamp: 0,
				}
				err := rows.Scan(&record.Opponent, &record.Win, &record.MyGesture, &record.OpGesture, &record.Timestamp)
				if err != nil {
					log.Fatal(err)
				} else {
					records = append(records, record)
				}
			}
			c.JSON(201, gin.H{
				"code":    201,
				"id":      id,
				"sid":     sid,
				"records": records,
			})
		} else {
			c.JSON(202, gin.H{
				"code":    202,
				"id":      id,
				"sid":     sid,
				"records": records,
			})
		}
	})
}

/*
	房间系统
	创建房间
	查看房间列表
	进入房间
	退出房间
	relativePath以/room作为开头
*/
func RoomSystem(r *gin.Engine) {
	RoomInsert(r)
	RoomList(r)
	RoomAll(r)
	RoomEnter(r)
	RoomLeave(r)
	RoomView(r)
}

/*
	个人创建房间
	请求类型:POST
	返回201说明创建成功
	返回202说明该账号创建房间失败
	后续relativePath为"/insert"
	Params中获取id和sid字段用以判断账号是否可以创建房间
*/
func RoomInsert(r *gin.Engine) {
	r.POST("/room/insert", func(c *gin.Context) {
		sidS := c.Query("sid")
		sid, _ := strconv.ParseInt(sidS, 10, 64)
		idS := c.Query("id")
		id, _ := strconv.ParseInt(idS, 10, 64)
		rid := int64(0)
		//验证该id和sid绑定的session是否存在
		if SessionCheck(sid, id, 0) {
			//创建房间并进入该房间
			rid = RoomCreate(id)
			c.JSON(201, gin.H{
				"code": 201,
				"sid":  sid,
				"id":   id,
				"rid":  rid,
			})
		} else {
			c.JSON(202, gin.H{
				"code": 202,
				"sid":  sid,
				"id":   id,
				"rid":  rid,
			})
		}
	})
}

/*
	查看可加入房间列表
	请求类型:GET
	返回201说明查看成功
	返回202说明认证错误
	后续relativePath为"/list"
	Params中获取id和sid字段用以判断账号是否可以查看房间列表
*/
func RoomList(r *gin.Engine) {
	r.GET("/room/list", func(c *gin.Context) {
		sidS := c.Query("sid")
		sid, _ := strconv.ParseInt(sidS, 10, 64)
		idS := c.Query("id")
		id, _ := strconv.ParseInt(idS, 10, 64)
		list := make([]Room, 0)
		//验证该id和sid绑定的session是否存在
		if SessionCheck(sid, id, 0) {
			for x := range RoomS {
				//仅加入二号玩家为空且房间编号不高于一万的房间号
				if RoomS[x].Player2 == 0 && RoomS[x].Room_id <= 10000 {
					list = append(list, RoomS[x])
				}
			}
			c.JSON(201, gin.H{
				"code": 201,
				"sid":  sid,
				"id":   id,
				"rooms":  list,
			})
		} else {
			c.JSON(202, gin.H{
				"code": 202,
				"sid":  sid,
				"id":   id,
				"rooms": list,
			})
		}
	})
}

/*
	查看所有房间列表
	请求类型:GET
	返回201说明查看成功
	返回202说明认证错误
	后续relativePath为"/all"
	Params中获取id和sid字段用以判断账号是否可以查看房间列表
*/
func RoomAll(r *gin.Engine) {
	r.GET("/room/all", func(c *gin.Context) {
		sidS := c.Query("sid")
		sid, _ := strconv.ParseInt(sidS, 10, 64)
		idS := c.Query("id")
		id, _ := strconv.ParseInt(idS, 10, 64)
		//对该id和sid绑定的session进行验证
		if SessionCheck(sid, id, 0) {
			c.JSON(201, gin.H{
				"code": 201,
				"sid":  sid,
				"id":   id,
				"rid":  RoomS,
			})
		} else {
			c.JSON(202, gin.H{
				"code": 202,
				"sid":  sid,
				"id":   id,
				"list": nil,
			})
		}
	})
}

/*
	进入房间
	请求类型:POST
	返回201说明进入成功
	返回202说明进入失败
	后续relativePath为"/enter"
	Params中获取id和sid字段用以认证账号处于登陆状态中
	Params中获取rid字段用以确定进入的房间号
*/
func RoomEnter(r *gin.Engine) {
	r.POST("/room/enter", func(c *gin.Context) {
		sidS := c.Query("sid")
		sid, _ := strconv.ParseInt(sidS, 10, 64)
		idS := c.Query("id")
		id, _ := strconv.ParseInt(idS, 10, 64)

		ridS := c.Query("rid")
		rid, _ := strconv.ParseInt(ridS, 10, 64)
		//验证账号处于登陆状态
		if SessionCheck(sid, id, 0) {
			if RoomEntrance(rid, id) {
				//进入所选的房间号并成为二号玩家
				c.JSON(201, gin.H{
					"code": 201,
					"sid":  sid,
					"id":   id,
					"rid":  rid,
				})
			} else {
				c.JSON(202, gin.H{
					"code": 202,
					"sid":  sid,
					"id":   id,
					"rid":  rid,
				})
			}
		} else {
			c.JSON(202, gin.H{
				"code": 202,
				"sid":  sid,
				"id":   id,
				"rid":  rid,
			})
		}
	})
}

/*
	进入观战
	请求类型:POST
	返回201说明进入成功
	返回202说明进入失败
	后续relativePath为"/view"
	Params中获取id和sid字段用以认证账号处于登陆状态中
	Params中获取rid字段用以确定进入观战的房间号
*/
func RoomView(r *gin.Engine) {
	r.POST("/room/view", func(c *gin.Context) {
		sidS := c.Query("sid")
		sid, _ := strconv.ParseInt(sidS, 10, 64)
		idS := c.Query("id")
		id, _ := strconv.ParseInt(idS, 10, 64)

		ridS := c.Query("rid")
		rid, _ := strconv.ParseInt(ridS, 10, 64)
		//验证该id和sid绑定的session存在
		if SessionCheck(sid, id, 0) {
			//进入房间编号为rid的房间进行观战
			if RoomWatch(rid, id) {
				c.JSON(201, gin.H{
					"code": 201,
					"sid":  sid,
					"id":   id,
					"rid":  rid,
				})
			} else {
				c.JSON(202, gin.H{
					"code": 202,
					"sid":  sid,
					"id":   id,
					"rid":  rid,
				})
			}
		} else {
			c.JSON(202, gin.H{
				"code": 202,
				"sid":  sid,
				"id":   id,
				"rid":  rid,
			})
		}
	})
}

/*
	离开房间
	请求类型:POST
	返回201说明退出成功
	返回202说明退出失败
	后续relativePath为"/leave"
	Params中获取id和sid字段用以认证账号处于在某一房间中
*/
func RoomLeave(r *gin.Engine) {
	r.POST("/room/leave", func(c *gin.Context) {
		sidS := c.Query("sid")
		sid, _ := strconv.ParseInt(sidS, 10, 64)
		idS := c.Query("id")
		id, _ := strconv.ParseInt(idS, 10, 64)
		ridS := c.Query("rid")
		rid, _ := strconv.ParseInt(ridS, 10, 64)
		//验证该账号处于登陆状态且session状态为在房间内(无论是否准备以及是否为观战)
		if SessionCheck(sid, id, 1) || SessionCheck(sid, id, 2) || SessionCheck(sid, id, 2975) {
			if RoomQuit(rid, id) {
				c.JSON(201, gin.H{
					"code": 201,
					"sid":  sid,
					"id":   id,
					"rid":  rid,
				})
			} else {
				c.JSON(202, gin.H{
					"code": 202,
					"sid":  sid,
					"id":   id,
					"rid":  rid,
				})
			}
		} else {
			c.JSON(202, gin.H{
				"code": 202,
				"sid":  sid,
				"id":   id,
				"rid":  rid,
			})
		}
	})
}

/*
	随机匹配
	请求类型:POST
	根据胜率分为三个等级
	胜率不高于0.33为低等级
	胜率高于0.33低于0.66为中等级
	胜率高于0.66为高等级
	按照三个等级进行分配使得匹配的对手和自己旗鼓相当
	relativePath为/random/match
*/
func RandomMatch(r *gin.Engine, db *sql.DB) {
	r.POST("/random/match", func(c *gin.Context) {
		sidS := c.Query("sid")
		sid, _ := strconv.ParseInt(sidS, 10, 64)
		idS := c.Query("id")
		id, _ := strconv.ParseInt(idS, 10, 64)
		rate := float32(0.0)
		rid := int64(0)
		//验证该账号处于登陆状态
		if SessionCheck(sid, id, 0) {
			//从数据库中查找该账号的胜率
			rows, _, _ := Database.SelectAllData(db, "account_list", "rate", "id="+idS)
			if rows.Next() {
				err := rows.Scan(&rate)
				if err != nil {
					c.JSON(202, gin.H{
						"code": 202,
						"sid":  sid,
						"id":   id,
						"rid":  rid,
					})
					log.Fatal(err)
				} else {
					//进入匹配模式
					rid = Random(id, rate)
					if rid != 0 {
						c.JSON(201, gin.H{
							"code": 201,
							"sid":  sid,
							"id":   id,
							"rid":  rid,
						})
					} else {
						c.JSON(202, gin.H{
							"code": 202,
							"sid":  sid,
							"id":   id,
							"rid":  rid,
						})
					}
				}
			}
		} else {
			c.JSON(202, gin.H{
				"code": 202,
				"sid":  sid,
				"id":   id,
				"rid":  rid,
			})
		}
	})
}

/*
	该房间游戏开始后用作倒计时
	倒计时时间为20s
	20s后判断当前房间的两位选手的手势
	如果某玩家中途掉线,其之前选择手势不变
	再次登陆时保持之前所选的手势
	即实现断线重连游戏不被迫终止
 */
func countDown(rid int64, db *sql.DB) {
	time.Sleep(20 * time.Second)
	var id1, id2 int64
	var gesture1, gesture2 int
	var win1, win2 int64
	var timestamp int64
	for x := range RoomS {
		if RoomS[x].Room_id == rid {
			id1 = RoomS[x].Player1
			id2 = RoomS[x].Player2
			gesture1 = RoomS[x].Prepare1 - 2
			gesture2 = RoomS[x].Prepare2 - 2
			RoomS[x].Prepare1 = 0
			RoomS[x].Prepare2 = 0
			if gesture1 == gesture2 {
				//平局
				win1 = 0
				win2 = 0
			} else {
				if gesture1 == 0 {
					win1 = 0
					win2 = 1
				} else if gesture2 == 0 {
					win1 = 1
					win2 = 0
				} else {
					if gesture1 == 1 {
						if gesture2 == 2 {
							win1 = 0
							win2 = 1
						} else {
							win1 = 1
							win2 = 0
						}
					} else if gesture1 == 2 {
						if gesture2 == 3 {
							win1 = 0
							win2 = 1
						} else {
							win1 = 1
							win2 = 0
						}
					} else if gesture1 == 3 {
						if gesture2 == 1 {
							win1 = 0
							win2 = 1
						} else {
							win1 = 1
							win2 = 0
						}
					}
				}
			}
			break
		}
	}
	var s string
	if win1 == 1 {
		s = "一号玩家获胜"
	} else if win2 == 1 {
		s = "二号玩家获胜"
	} else {
		s = "双方平局"
	}
	for x := range Monitors {
		if Monitors[x].rid == rid {
			fmt.Fprintf(Monitors[x].conn, s+string('\n'))
		}
	}

	timestamp = time.Now().Unix()
	id1S := strconv.FormatInt(id1, 10)
	id2S := strconv.FormatInt(id2, 10)
	gesture1S := strconv.Itoa(gesture1)
	gesture2S := strconv.Itoa(gesture2)
	win1S := strconv.FormatInt(win1, 10)
	win2S := strconv.FormatInt(win2, 10)
	timestampS := strconv.FormatInt(timestamp, 10)
	Database.InsertData(db, "record"+id1S, id2S+","+win1S+","+gesture1S+","+gesture2S+","+timestampS)
	Database.InsertData(db, "record"+id2S, id1S+","+win2S+","+gesture2S+","+gesture1S+","+timestampS)
	var rate1, rate2 float32
	var total1, total2 int64
	total1 = Database.SelectKeyGetFieldI(db, "account_list", "total", "id="+id1S) + 1
	total2 = Database.SelectKeyGetFieldI(db, "account_list", "total", "id="+id2S) + 1
	total1S := strconv.FormatInt(total1, 10)
	total2S := strconv.FormatInt(total2, 10)
	win1 = win1 + Database.SelectKeyGetFieldI(db, "account_list", "win", "id="+id1S)
	win2 = win2 + Database.SelectKeyGetFieldI(db, "account_list", "win", "id="+id2S)
	win1S = strconv.FormatInt(win1, 10)
	win2S = strconv.FormatInt(win2, 10)
	rate1 = float32(win1) / float32(total1)
	rate2 = float32(win2) / float32(total2)
	rate1S := fmt.Sprintf("%.2f", rate1)
	rate2S := fmt.Sprintf("%.2f", rate2)
	Database.UpdateData(db, "account_list", "rate="+rate1S+",win="+win1S+",total="+total1S, "id="+id1S)
	Database.UpdateData(db, "account_list", "rate="+rate2S+",win="+win2S+",total="+total2S, "id="+id2S)
}

/*
	每一个客户端监听的携程
	用于接受处理每一条该客户端发来的消息
 */
func GetMessageFromCustomer(conn net.Conn, db *sql.DB) {
	for {
		buf := make([]byte, 1024)
		n, err := conn.Read(buf) //从coon读取
		if err != nil {
			fmt.Printf("客户端%s退出\n", conn.RemoteAddr().String())
			for x := 0; x < len(Monitors); x++ {
				if Monitors[x].conn == conn {
					Monitors = append(Monitors[:x], Monitors[x+1:]...)
					return
				}
			}
			return
		}
		mess := message{
			Mes: "",
			Id:  0,
			Sid: 0,
			Rid: 0,
		}
		json.Unmarshal([]byte(string(buf[:n])), &mess)
		if mess.Mes == "1" {
			//准备消息
			s := ""
			for x := range RoomS {
				if RoomS[x].Room_id == mess.Rid {
					if RoomS[x].Player1 == mess.Id {
						RoomS[x].Prepare1 = 1
						s = "一号选手已准备"
					}
					if RoomS[x].Player2 == mess.Id {
						RoomS[x].Prepare2 = 1
						s = "二号选手已准备"
					}
					break
				}
			}
			for x := range Monitors {
				if Monitors[x].rid == mess.Rid {
					fmt.Fprintf(Monitors[x].conn, s+string('\n'))
				}
			}
			for x := range RoomS {
				if RoomS[x].Room_id == mess.Rid {
					if RoomS[x].Prepare1 == 1 && RoomS[x].Prepare2 == 1 {
						//进入倒计时状态
						RoomS[x].Prepare1 = 2
						RoomS[x].Prepare2 = 2
						go countDown(mess.Rid, db)
						//通知所有处于该房间的玩家游戏已开始
						for z := range Monitors {
							if Monitors[z].rid == mess.Rid {
								fmt.Fprintf(Monitors[z].conn, "两位选手已经准备就绪,游戏开始"+string('\n'))
							}
						}
					}
					break
				}
			}
		} else if mess.Mes == "3" || mess.Mes == "4" || mess.Mes == "5" {
			//选择手势,若游戏未开始则无效
			var gesture int
			if mess.Mes == "3" {
				gesture = 3
			} else if mess.Mes == "4" {
				gesture = 4
			} else if mess.Mes == "5" {
				gesture = 5
			}
			for x := range RoomS {
				if RoomS[x].Room_id == mess.Rid {
					if RoomS[x].Prepare1 >= 2 && RoomS[x].Prepare2 >= 2 {
						if RoomS[x].Player1 == mess.Id {
							RoomS[x].Prepare1 = gesture
						}
						if RoomS[x].Player2 == mess.Id {
							RoomS[x].Prepare2 = gesture
						}
					}
					break
				}
			}
		} else {
			//聊天消息,所有处于该房间内的人都可以看到
			b, _ := json.Marshal(mess)
			s := string(b)
			for x := range Monitors {
				if Monitors[x].rid == mess.Rid {
					fmt.Fprintf(Monitors[x].conn, s+string('\n'))
				}
			}
		}
		if mess.Mes == "exit" {
			break
		}
	}
}

/*
	socket
	监听端口2997
	对于每个监听的对象有要求发来对应的id,sid和房间号rid
	其中,id和sid用于验证该账号处于登陆状态中
	rid用于在后续消息分发国产中判断接受某一房间的消息
 */
func socket(db *sql.DB) {
	l, err := net.Listen("tcp", ":2997")
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}
		defer conn.Close()
		buf := make([]byte, 1024)
		//接受第一条消息用于判断该链接的情况
		n, _ := conn.Read(buf)
		mon := monitor{
			conn: conn,
			id:   0,
			sid:  0,
			rid:  0,
		}
		var mess message
		json.Unmarshal([]byte(string(buf[:n])), &mess)
		//判断该链接绑定的账号是否处于房间内或观战模式
		if SessionCheck(mess.Sid, mess.Id, 1) || SessionCheck(mess.Sid, mess.Id, 2975) {
			mon.id = mess.Id
			mon.sid = mess.Sid
			mon.rid = mess.Rid
			Monitors = append(Monitors, mon)
			go GetMessageFromCustomer(conn, db)
		} else {
			fmt.Fprintf(conn, "error"+string('\n'))
			fmt.Println("error")
			conn.Close()
		}
	}
}
func main() {
	r := gin.Default()
	_, _, db := CreateList()
	//开一个协程用以监听所有处于房间中的链接的消息
	go socket(db)
	AccountSystem(r, db)
	RoomSystem(r)
	RandomMatch(r, db)
	r.Run(":2995")
}
