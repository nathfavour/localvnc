package server

import (
	"fmt"
	"github.com/nathfavour/localvnc/screen"
	"net/http"
	"time"
)

func Start(port int) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "multipart/x-mixed-replace; boundary=frame")
		for {
			img, err := screen.CapturePNG()
			if err != nil {
				fmt.Println("Capture error:", err)
				break
			}
			fmt.Fprintf(w, "--frame\r\nContent-Type: image/png\r\n\r\n")
			w.Write(img)
			fmt.Fprintf(w, "\r\n")
			time.Sleep(100 * time.Millisecond)
		}
	})
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
