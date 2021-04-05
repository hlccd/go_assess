package IM

import (
	"strings"
)

type Message struct {
	Id    int64  `json:"id"`
	Group int64  `json:"group"`
	Mes   string `json:"mes"`
}

func SToCanSend(s string) string {
	s = strings.Replace(s, string('\n'), "|+|", -1)
	return s
}
func SToCanRead(s string) string {
	s = strings.Replace(s, "|+|", string('\n'), -1)
	return s
}

