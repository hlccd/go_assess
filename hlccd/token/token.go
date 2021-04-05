package token

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
)

type Header struct {
	Alg string `json:"alg"`
}

type Payload struct {
	Iss      string `json:"iss"`
	Exp      string `json:"exp"`
	Iat      string `json:"iat"`
	UserID   string `json:"userID"`
}

//CreateHeader 创建Token头部
func CreateHeader() string {
	h := Header{
		Alg: "HS256",
	}
	bytes, err := json.Marshal(h)
	if err != nil {
		log.Printf("marshal failed!序列化失败！错误信息：%v", err)
		return ""
	}
	header := base64.StdEncoding.EncodeToString(bytes)
	return header
}

//CreatePayload 创建Token载荷
func CreatePayload(userID string) string {
	p := Payload{
		Iss:      "hlccd",
		Exp:      strconv.FormatInt(time.Now().Add(15*24*time.Hour).Unix(), 10),
		Iat:      strconv.FormatInt(time.Now().Unix(), 10),
		UserID:   userID,
	}
	bytes, err := json.Marshal(p)
	if err != nil {
		log.Printf("marshal failed!序列化失败！错误信息：%v", err)
		return ""
	}
	payload := base64.StdEncoding.EncodeToString(bytes)
	return payload
}

//CreateSignature 创建签证
func CreateSignature(userID string) string {
	h := CreateHeader()
	p := CreatePayload(userID)
	str := strings.Join([]string{h, p}, ".")
	key := "2975hLcCd"
	hash := hmac.New(sha256.New, []byte(key))
	hash.Write([]byte(str))
	s := hash.Sum(nil)
	signature := base64.StdEncoding.EncodeToString(s)
	return signature
}

//Create 登陆后创建Token
func CreateToken(userID string) (token string) {
	header := CreateHeader()
	payload := CreatePayload(userID)
	signature := CreateSignature(userID)
	token = strings.Join([]string{header, payload}, ".") + "." + signature
	return token
}

//CheckToken 核对检查token头部、载荷、签证三部分信息
func CheckToken(token string) (payload Payload,err error) {
	token=strings.Replace(token, " ", "+", -1)
	split := strings.Split(token, ".")
	if len(split) != 3 {
		err := errors.New("token构建错误")
		log.Println(err)
		return payload,err
	}
	fmt.Println(split)
	_, err = base64.StdEncoding.DecodeString(split[0])
	if err != nil {
		err = errors.New("header解析错误")
		log.Println(err)
		return payload,err
	}
	p, err := base64.StdEncoding.DecodeString(split[1])
	if err != nil {
		err = errors.New("payload解析错误")
		log.Println(err)
		return payload,err
	}
	_, err = base64.StdEncoding.DecodeString(split[2])
	if err != nil {
		err = errors.New("signature解析错误")
		log.Println(err)
		return payload,err
	}
	err = json.Unmarshal(p, &payload)
	if err != nil {
		log.Printf("unmarshal failed!反序列化失败！错误信息：%v", err)
		return payload,err
	}
	return payload,nil
}