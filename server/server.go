package server

import (
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/nathfavour/localvnc/screen"
)

func Start(port int) error {
	var latestFrame atomic.Value

	// Frame capture goroutine
	go func() {
		for {
			img, err := screen.CaptureJPEG()
			if err == nil {
				latestFrame.Store(img)
			}
			time.Sleep(100 * time.Millisecond)
		}
	}()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "multipart/x-mixed-replace; boundary=frame")
		for {
			frame := latestFrame.Load()
			if frame == nil {
				time.Sleep(10 * time.Millisecond)
				continue
			}
			img := frame.([]byte)
			fmt.Fprintf(w, "--frame\r\nContent-Type: image/jpeg\r\n\r\n")
			w.Write(img)
			fmt.Fprintf(w, "\r\n")
			time.Sleep(100 * time.Millisecond)
		}
	})
	return http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
