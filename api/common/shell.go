package common

import (
	"log"
	"net"
	"os"
	"strings"
)

func GetLocalIp() string {
	conn, err := net.Dial("udp", "8.8.8.8:53")
	defer conn.Close()
	if err != nil {
		log.Printf("get local addr err:%v", err)
		return ""
	}
	localIp := strings.Split(conn.LocalAddr().String(), ":")[0]
	return localIp
}

func GetHostName() string {
	name, _ := os.Hostname()
	return name
}
