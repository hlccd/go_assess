package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strings"
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
	}
}
func main() {
	//输入开始要绑定的身份信息
	fmt.Println("请以此输入id,sid,rid:")
	var id,sid,rid int64
	fmt.Scanln(&id,&sid,&rid)
	conn, err := net.Dial("tcp", "47.108.217.244:2997")
	if err != nil {
		log.Fatal(err)
	}
	mes := message{
		Mes: "建立链接",
		Id:  id,
		Sid: sid,
		Rid: rid,
	}
	b, err := json.Marshal(mes)
	if err != nil {
		fmt.Println("JSON ERR:", err)
	}
	//建立链接的同时发送消息用于绑定身份
	fmt.Fprintf(conn, string(b)+string('\n'))
	defer conn.Close()
	go getS(conn)
	var s string
	fmt.Println("请在下方输入指令")
	for true {
		//读取客户端本地的消息并转化成json格式发送给服务器
		fmt.Scanln(&s)
		mes := message{
			Mes: s,
			Id:  id,
			Sid: sid,
			Rid: rid,
		}
		b, err := json.Marshal(mes)
		if err != nil {
			fmt.Println("JSON ERR:", err)
		}
		fmt.Fprintf(conn, string(b)+string('\n'))
	}
}
