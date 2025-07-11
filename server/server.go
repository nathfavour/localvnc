package server

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/BurntSushi/xgb/xproto"

	"github.com/nathfavour/localvnc/screen"
)

var embeddedClientHTML = `<!DOCTYPE html>
<html>
<head>
  <title>LocalVNC Advanced Client</title>
  <style>
    body { background: #222; color: #eee; font-family: sans-serif; }
    canvas { border: 2px solid #444; background: #000; }
    #controls { margin: 10px 0; }
  </style>
</head>
<body>
  <h2>LocalVNC Delta Frame Client</h2>
  <div id="controls">
    <button onclick="clearCanvas()">Clear Canvas</button>
    <span id="status"></span>
  </div>
  <canvas id="screen"></canvas>
  <script>
    const serverHost = window.location.hostname;
    const serverPort = window.location.port || 3456;
    const screenSizeEndpoint = "/screen-size";
    const deltaEndpoint = "/";
    let canvas = document.getElementById('screen');
    let ctx = canvas.getContext('2d');
    let status = document.getElementById('status');
    let screenWidth = 1920, screenHeight = 1080;
    fetch(screenSizeEndpoint)
      .then(res => res.json())
      .then(size => {
        screenWidth = size.width;
        screenHeight = size.height;
        canvas.width = screenWidth;
        canvas.height = screenHeight;
        status.textContent = "Screen size: " + screenWidth + "x" + screenHeight;
        poll();
      })
      .catch(() => {
        canvas.width = screenWidth;
        canvas.height = screenHeight;
        status.textContent = "Using default screen size: " + screenWidth + "x" + screenHeight;
        poll();
      });
    function clearCanvas() {
      ctx.clearRect(0, 0, canvas.width, canvas.height);
    }
    function patchRegion(x, y, w, h, jpegData) {
      const img = new Image();
      img.onload = function() {
        ctx.drawImage(img, x, y, w, h);
      };
      img.src = 'data:image/jpeg;base64,' + jpegData;
    }
    function poll() {
      fetch(deltaEndpoint)
        .then(res => res.json())
        .then(deltas => {
          if (Array.isArray(deltas)) {
            deltas.forEach(d => {
              patchRegion(d.X, d.Y, d.Width, d.Height, d.Data);
            });
            status.textContent = "Received " + deltas.length + " region(s)";
          } else {
            status.textContent = "No regions received";
          }
          setTimeout(poll, 30);
        })
        .catch(() => {
          status.textContent = "Connection lost, retrying...";
          setTimeout(poll, 500);
        });
    }
  </script>
</body>
</html>
`

func Start(port int) error {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		enc := json.NewEncoder(w)
		for {
			deltas, err := screen.CaptureDeltaJPEG()
			if err != nil {
				http.Error(w, fmt.Sprintf("Screen capture error: %v", err), http.StatusInternalServerError)
				fmt.Println("Capture error:", err)
				break
			}
			if len(deltas) == 0 {
				time.Sleep(50 * time.Millisecond)
				continue
			}
			// Encode each delta frame's Data as base64 for JSON transport
			type DeltaOut struct {
				X, Y, Width, Height int
				Data                string
			}
			var out []DeltaOut
			for _, d := range deltas {
				out = append(out, DeltaOut{
					X:      d.X,
					Y:      d.Y,
					Width:  d.Width,
					Height: d.Height,
					Data:   base64.StdEncoding.EncodeToString(d.Data),
				})
			}
			enc.Encode(out)
			time.Sleep(50 * time.Millisecond)
		}
	})

	http.HandleFunc("/screen-size", func(w http.ResponseWriter, r *http.Request) {
		conn, err := screen.GetConn()
		if err != nil {
			http.Error(w, "Could not get screen size", http.StatusInternalServerError)
			return
		}
		setup := xproto.Setup(conn)
		s := setup.DefaultScreen(conn)
		size := struct {
			Width  int `json:"width"`
			Height int `json:"height"`
		}{
			Width:  int(s.WidthInPixels),
			Height: int(s.HeightInPixels),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(size)
	})

	http.HandleFunc("/client", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, embeddedClientHTML)
	})

	return http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
