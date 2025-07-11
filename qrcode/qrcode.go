package qrcode

import (
	"fmt"
	"github.com/skip2/go-qrcode"
)

func PrintQRCode(url string) {
	qr, _ := qrcode.New(url, qrcode.Medium)
	fmt.Println(qr.ToString(false))
}
