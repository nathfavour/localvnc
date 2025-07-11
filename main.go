package main

import (
	"fmt"
	"os"
	"net"
	"github.com/nathfavour/localvnc/screen"
	"github.com/nathfavour/localvnc/server"
	"github.com/nathfavour/localvnc/qrcode"
)

func main() {
	port := 3456
	// Get local IP address
	addrs, _ := net.InterfaceAddrs()
	var localIP string
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
			localIP = ipnet.IP.String()
			break
		}
	}
	url := fmt.Sprintf("http://%s:%d", localIP, port)
	fmt.Println("Streaming on:", url)
	qrcode.PrintQRCode(url)
	go server.Start(port)
	select {} // Block forever
}
package localvnc
