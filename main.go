package main

import (
	"flag"
	"fmt"
	"net"
	"os"

	"github.com/nathfavour/localvnc/qrcode"
	"github.com/nathfavour/localvnc/server"
)

func main() {
	// Add port and help flags
	port := flag.Int("port", 3456, "Port to run the server on")
	help := flag.Bool("help", false, "Show help message")
	flag.Parse()

	if *help {
		fmt.Println("Usage: localvnc [--port PORT]")
		flag.PrintDefaults()
		os.Exit(0)
	}

	// Get local IP address
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting network interfaces: %v\n", err)
		os.Exit(1)
	}
	var localIP string
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
			localIP = ipnet.IP.String()
			break
		}
	}
	if localIP == "" {
		fmt.Fprintln(os.Stderr, "Could not determine local IP address. Is your network connected?")
		os.Exit(1)
	}
	url := fmt.Sprintf("http://%s:%d", localIP, *port)
	fmt.Println("Streaming on:", url)
	qrcode.PrintQRCode(url)

	fmt.Printf("Starting server on port %d...\n", *port)
	err = server.Start(*port)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start server: %v\n", err)
		fmt.Fprintln(os.Stderr, "If you see a blank page, check your system's screen capture permissions.")
		os.Exit(1)
	}
	select {} // Block forever
}
