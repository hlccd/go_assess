package IM

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
)


func GetConn(address string) (net.Conn,error){
	conn, err := net.Dial("tcp", address)
	if err != nil {
		fmt.Println("获取Conn失败")
		log.Fatal(err)
		return nil,err
	}
	return conn,nil
}

func SendMessageToServer(conn net.Conn,id int64,group int64,s string) {
	mes:=Message{
		Id:    id,
		Group: group,
		Mes:   SToCanRead(s),
	}
	b, err := json.Marshal(mes)
	if err != nil {
		fmt.Println("JSON ERR:", err)
	}
	fmt.Fprintf(conn, string(b)+string('\n'))
}

func GetMessageFromServer(conn net.Conn) {
	for true {
		data, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		data=SToCanRead(data)
		var mes Message
		json.Unmarshal([]byte(data), &mes)
		fmt.Printf(string(data))
		fmt.Println(mes.Mes)
	}
}

