package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

/*
	消息结构
	与服务端消息结构相同
	描述见readme
*/
type message struct {
	Mes string `json:"mes"`
	Id  int64  `json:"id"`
	Sid int64  `json:"sid"`
	Rid int64  `json:"rid"`
}

//用户结构
type user struct {
	Code     int    `json:"code"`
	Id       int64  `json:"id"`
	Password string `json:"password"`
	Sid      int64  `json:"sid"`
	Rid      int64  `json:"rid"`
}

var User user

/*
	用于承接反馈json的结构体
*/

/*
	字符串转化
	暂未使用
*/
func sToCanSend(s string) string {
	s = strings.Replace(s, string('\n'), "|+|", -1)
	return s
}

/*
	字符串转可读
*/
func sToCanRead(s string) string {
	s = strings.Replace(s, "|+|", string('\n'), -1)
	return s
}

/*
	开一个可获取到服务器发来的消息的携程
	将消息转化为可存储于message的内容并打印
	同时打印消息主体
*/
func getS(conn net.Conn) {
	for true {
		data, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		data = sToCanRead(data)
		var mes message
		json.Unmarshal([]byte(data), &mes)
		fmt.Printf(string(data))
		fmt.Println(mes.Mes)
		if mes.Mes == "exit" {
			break
		}
	}
}

func menu(url1 string) {
	var num int
	fmt.Println("输入指令代码做出选择:")
	fmt.Println("1:注册")
	fmt.Println("2:登陆")
	fmt.Scan(&num)
	if num == 1 {
		register(url1)
	} else if num == 2 {
		login(url1)
	} else {
		menu(url1)
	}
}
func register(url1 string) {
	url := url1 + "/account/register?"
	fmt.Println("请输入密码:")
	password := ""
	fmt.Scan(&password)
	url += "password=" + password
	resp, _ := http.Post(url,
		"application/x-www-form-urlencoded",
		strings.NewReader("name=cjb"))
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	json.Unmarshal([]byte(string(body)), &User)
	fmt.Println(string(body))
	if User.Code == 201 {
		fmt.Println("注册成功")
		fmt.Println(User)
	} else {
		fmt.Println("注册失败")
	}
	menu(url1)
}
func login(url1 string) {
	id := int64(0)
	password := ""
	fmt.Print("请输入ID号:")
	fmt.Scan(&id)
	fmt.Print("请输入密码:")
	fmt.Scan(&password)
	url := url1 + "/account/login?"
	url += "password=" + password + "&id=" + strconv.FormatInt(id, 10)
	resp, _ := http.Post(url,
		"application/x-www-form-urlencoded",
		strings.NewReader("name=cjb"))
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	json.Unmarshal([]byte(string(body)), &User)
	if User.Code == 201 {
		fmt.Println("登陆成功")
		fmt.Println(string(body))
		fmt.Println(User)
		accountMenu(url1)
	} else {
		fmt.Println("登陆失败")
		menu(url1)
	}
}
func accountMenu(url1 string) {
	var num int
	fmt.Println("输入指令代码做出选择:")
	fmt.Println("1:登出")
	fmt.Println("2:查看历史对局")
	fmt.Println("3:匹配")
	fmt.Println("4:创建房间")
	fmt.Println("5:查看可加入房间列表")
	fmt.Println("6:查看全部房间列表")
	fmt.Println("7:进入房间")
	fmt.Println("8:进入房间观战")
	fmt.Println(User)
	fmt.Scan(&num)
	if num == 1 {
		logout(url1)
	} else if num == 2 {
		record(url1)
	}else if num==3{
		match(url1)
	}else if num==4{
		insert(url1)
	}else if num == 5 {
		list(url1)
	}else if num==6{
		all(url1)
	}else if num==7{
		enter(url1)
	}else if num==8{
		view(url1)
	}else {
		accountMenu(url1)
	}
}
func logout(url1 string) {
	url := url1 + "/account/logout?"
	url += "sid=" + strconv.FormatInt(User.Sid, 10) + "&id=" + strconv.FormatInt(User.Id, 10)
	resp, _ := http.Post(url,
		"application/x-www-form-urlencoded",
		strings.NewReader("name=cjb"))
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	json.Unmarshal([]byte(string(body)), &User)
	if User.Code == 201 {
		fmt.Println("登出成功")
		fmt.Println(string(body))
		User = user{
			Code:     0,
			Id:       0,
			Password: "",
			Sid:      0,
			Rid:      0,
		}
		fmt.Println(User)
	} else {
		fmt.Println("登出失败")
	}
	menu(url1)
}
func record(url1 string) {
	url := url1 + "/account/record?"
	url += "sid=" + strconv.FormatInt(User.Sid, 10) + "&id=" + strconv.FormatInt(User.Id, 10)
	resp, _ := http.Get(url)
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	json.Unmarshal([]byte(string(body)), &User)
	if User.Code == 201 {
		fmt.Println("查看历史对局记录成功")
		fmt.Println(string(body))
		fmt.Println(User)
	} else {
		fmt.Println("查看历史对局记录失败")
	}
	accountMenu(url1)
}
func match(url1 string) {
	url := url1 + "/random/match?"
	url += "sid=" + strconv.FormatInt(User.Sid, 10) + "&id=" + strconv.FormatInt(User.Id, 10)
	resp, _ := http.Post(url,
		"application/x-www-form-urlencoded",
		strings.NewReader("name=cjb"))
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	json.Unmarshal([]byte(string(body)), &User)
	if User.Code==201{
		fmt.Println("进入匹配成功")
		fmt.Println(string(body))
		socket()
		leave(url1)
	}else {
		fmt.Println("进入匹配失败")
		accountMenu(url1)
	}
}
func insert(url1 string)  {
	url := url1 + "/room/insert?"
	url += "sid=" + strconv.FormatInt(User.Sid, 10) + "&id=" + strconv.FormatInt(User.Id, 10)
	resp, _ := http.Post(url,
		"application/x-www-form-urlencoded",
		strings.NewReader("name=cjb"))
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	json.Unmarshal([]byte(string(body)), &User)
	if User.Code==201 {
		fmt.Println("创建房间成功,房间编号:"+strconv.FormatInt(User.Rid,10))
		fmt.Println(string(body))
		fmt.Println(User)
		socket()
		leave(url1)
	}else {
		fmt.Println("创建房间失败")
		accountMenu(url1)
	}
}
func enter(url1 string) {
	var rid int64
	fmt.Println("请输入要加入的房间编号")
	fmt.Scan(&rid)
	url := url1 + "/room/enter?"
	url += "sid=" + strconv.FormatInt(User.Sid, 10) + "&id=" + strconv.FormatInt(User.Id, 10)+"&rid="+strconv.FormatInt(rid,10)
	resp, _ := http.Post(url,
		"application/x-www-form-urlencoded",
		strings.NewReader("name=cjb"))
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	json.Unmarshal([]byte(string(body)), &User)
	if User.Code==201 {
		fmt.Println("进入房间成功")
		fmt.Println(string(body))
		fmt.Println(User)
		socket()
		leave(url1)
	}else{
		fmt.Println("进入房间失败")
		accountMenu(url1)
	}
}
func list(url1 string) {
	url := url1 + "/room/list?"
	url += "sid=" + strconv.FormatInt(User.Sid, 10) + "&id=" + strconv.FormatInt(User.Id, 10)
	resp, _ := http.Get(url)
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	json.Unmarshal([]byte(string(body)), &User)
	if User.Code == 201 {
		fmt.Println("查看可进入房间成功")
		fmt.Println(string(body))
		fmt.Println(User)
	} else {
		fmt.Println("查看可进入房间失败")
	}
	accountMenu(url1)
}
func all(url1 string) {
	url := url1 + "/room/all?"
	url += "sid=" + strconv.FormatInt(User.Sid, 10) + "&id=" + strconv.FormatInt(User.Id, 10)
	resp, _ := http.Get(url)
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	json.Unmarshal([]byte(string(body)), &User)
	if User.Code == 201 {
		fmt.Println("查看所有房间成功")
		fmt.Println(string(body))
		fmt.Println(User)
	} else {
		fmt.Println("查看所有房间失败")
	}
	accountMenu(url1)
}
func view(url1 string) {
	var rid int64
	fmt.Println("请输入要观战的房间编号")
	fmt.Scan(&rid)
	url := url1 + "/room/view?"
	url += "sid=" + strconv.FormatInt(User.Sid, 10) + "&id=" + strconv.FormatInt(User.Id, 10)+"&rid="+strconv.FormatInt(rid,10)
	resp, _ := http.Post(url,
		"application/x-www-form-urlencoded",
		strings.NewReader("name=cjb"))
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	json.Unmarshal([]byte(string(body)), &User)
	if User.Code==201 {
		fmt.Println("进入房间观战成功")
		fmt.Println(string(body))
		fmt.Println(User)
		socket()
		leave(url1)
	}else{
		fmt.Println("进入房间观战失败")
		accountMenu(url1)
	}
}


func socket() {
	//输入开始要绑定的身份信息
	conn, _ := net.Dial("tcp", "47.108.217.244:2997")
	mes := message{
		Mes: "建立链接",
		Id:  User.Id,
		Sid: User.Sid,
		Rid: User.Rid,
	}
	b, err := json.Marshal(mes)
	if err != nil {
		fmt.Println("JSON ERR:", err)
	}
	//建立链接的同时发送消息用于绑定身份
	fmt.Fprintf(conn, string(b)+string('\n'))
	go getS(conn)
	var s string
	fmt.Println("请在下方输入指令,若指令为'exit'则表示退出该房间")
	for true {
		//读取客户端本地的消息并转化成json格式发送给服务器
		fmt.Scanln(&s)
		mes := message{
			Mes: s,
			Id:  User.Id,
			Sid: User.Sid,
			Rid: User.Rid,
		}
		b, err := json.Marshal(mes)
		if err != nil {
			fmt.Println("JSON ERR:", err)
		}
		fmt.Fprintf(conn, string(b)+string('\n'))
		if s=="exit"{
			time.Sleep(1*time.Second)
			break
		}
	}
}
func leave(url1 string) {
	url := url1 + "/room/leave?"
	url += "sid=" + strconv.FormatInt(User.Sid, 10) + "&id=" + strconv.FormatInt(User.Id, 10)+"&rid="+strconv.FormatInt(User.Rid,10)
	resp, _ := http.Post(url,
		"application/x-www-form-urlencoded",
		strings.NewReader("name=cjb"))
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	json.Unmarshal([]byte(string(body)), &User)
	if User.Code==201{
		fmt.Println("退出房间成功")
		fmt.Println(string(body))
	}else {
		fmt.Println("退出房间失败")
	}
	accountMenu(url1)
}
func main() {
	url1 := "http://47.108.217.244:2995"
	menu(url1)
}
