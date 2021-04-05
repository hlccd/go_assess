package IM

import (
	"fmt"
	"net"
	"log"
)

var List = make([]net.Conn, 0)

func GetListen(address string) (net.Listener, error) {
	l, err := net.Listen("tcp", address)
	if err != nil {
		fmt.Println("获取Listener失败")
		log.Fatal(err)
		return nil, err
	}
	return l, nil
}
func ListAdd(conn net.Conn) {
	List = append(List, conn)
}
func ListDelete(x int) int {
	List = append(List[:x], List[x+1:]...)
	return x - 1
}
func HandleConnection(conn net.Conn) {
	go GetMessageFromCustomer(conn)
}
func GetMessageFromCustomer(conn net.Conn) {
	for {
		//创建一个新的切片
		buf := make([]byte, 1024)
		//coon.Read(buf)
		//1.等待客户端通过coon发送消息
		//2.客户端没有写，则协程阻塞在这儿
		//fmt.Printf("服务器在等待客户端%s发送信息\n",conn.RemoteAddr().String())
		n, err := conn.Read(buf) //从coon读取
		if err != nil {
			fmt.Printf("客户端%s退出\n", conn.RemoteAddr().String())
			for x := 0; x < len(List); x++ {
				if List[x] == conn {
					x = ListDelete(x)
				}
			}
			return
		}
		for x := range List {
			fmt.Fprintf(List[x], string(buf[:n])+string('\n'))
		}
		fmt.Println(string(buf[:n]))
	}
}
