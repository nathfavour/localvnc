package screen

import (
	"bytes"
	"image"
	"image/jpeg"
	"image/png"
	"math/rand"
	"sort"
	"sync"

	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/damage"
	"github.com/BurntSushi/xgb/xproto"
	"golang.org/x/image/draw"
)

var (
	conn     *xgb.Conn
	connOnce sync.Once
	damageId damage.Damage
)

func getConn() (*xgb.Conn, error) {
	var err error
	connOnce.Do(func() {
		conn, err = xgb.NewConn()
	})
	return conn, err
}

// Exported for use in other packages
func GetConn() (*xgb.Conn, error) {
	return getConn()
}

// Initialize XDamage and create a damage object for the root window
func initXDamage() error {
	c, err := getConn()
	if err != nil {
		return err
	}
	err = damage.Init(c)
	if err != nil {
		return err
	}
	setup := xproto.Setup(c)
	screen := setup.DefaultScreen(c)
	root := screen.Root
	// Generate a unique ID for the Damage object (xgb does not provide NewId, so use random uint32)
	damageId = damage.Damage(uint32(rand.Int31()))

	const reportLevelNonEmpty = 1
	cookie := damage.CreateChecked(c, damageId, xproto.Drawable(root), reportLevelNonEmpty)
	if err := cookie.Check(); err != nil {
		return err
	}
	return nil
}

// Poll for XDamage events and return changed rectangles
func GetChangedRects() ([]xproto.Rectangle, error) {
	c, err := getConn()
	if err != nil {
		return nil, err
	}
	if damageId == 0 {
		if err := initXDamage(); err != nil {
			return nil, err
		}
	}
	var rects []xproto.Rectangle
	for {
		ev, err := c.WaitForEvent()
		if err != nil || ev == nil {
			break
		}
		switch e := ev.(type) {
		case damage.NotifyEvent:
			rects = append(rects, e.Area)
			return rects, nil
		default:
			// Ignore other events
			return rects, nil
		}
	}
	return rects, nil
}

func CapturePNG() ([]byte, error) {
	conn, err := getConn()
	if err != nil {
		return nil, err
	}

	setup := xproto.Setup(conn)
	screen := setup.DefaultScreen(conn)
	width := int(screen.WidthInPixels)
	height := int(screen.HeightInPixels)

	reply, err := xproto.GetImage(conn, xproto.ImageFormatZPixmap, xproto.Drawable(screen.Root),
		0, 0, uint16(width), uint16(height), 0xffffffff).Reply()
	if err != nil {
		return nil, err
	}

	// Convert BGRA (X11) to RGBA (Go)
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	src := reply.Data
	dst := img.Pix
	for i := 0; i+3 < len(src) && i+3 < len(dst); i += 4 {
		// X11: B,G,R,X; Go: R,G,B,A
		dst[i+0] = src[i+2] // R
		dst[i+1] = src[i+1] // G
		dst[i+2] = src[i+0] // B
		dst[i+3] = 255      // A
	}

	// Downscale to 3/4 size
	newWidth := width * 3 / 4
	newHeight := height * 3 / 4
	downscaled := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
	draw.ApproxBiLinear.Scale(downscaled, downscaled.Rect, img, img.Bounds(), draw.Over, nil)

	buf := new(bytes.Buffer)
	err = png.Encode(buf, downscaled)
	return buf.Bytes(), err
}

func CaptureJPEG() ([]byte, error) {
	conn, err := getConn()
	if err != nil {
		return nil, err
	}

	setup := xproto.Setup(conn)
	screen := setup.DefaultScreen(conn)
	width := int(screen.WidthInPixels)
	height := int(screen.HeightInPixels)

	reply, err := xproto.GetImage(conn, xproto.ImageFormatZPixmap, xproto.Drawable(screen.Root),
		0, 0, uint16(width), uint16(height), 0xffffffff).Reply()
	if err != nil {
		return nil, err
	}

	// Convert BGRA (X11) to RGBA (Go)
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	src := reply.Data
	dst := img.Pix
	for i := 0; i+3 < len(src) && i+3 < len(dst); i += 4 {
		dst[i+0] = src[i+2] // R
		dst[i+1] = src[i+1] // G
		dst[i+2] = src[i+0] // B
		dst[i+3] = 255      // A
	}

	// Downscale to 3/4 size
	newWidth := width * 3 / 4
	newHeight := height * 3 / 4
	downscaled := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
	draw.ApproxBiLinear.Scale(downscaled, downscaled.Rect, img, img.Bounds(), draw.Over, nil)

	buf := new(bytes.Buffer)
	// Use jpeg encoding with quality 80 (adjust as needed)
	err = jpeg.Encode(buf, downscaled, &jpeg.Options{Quality: 80})
	return buf.Bytes(), err
}

type DeltaFrame struct {
	X, Y, Width, Height int
	Data                []byte
}

// Merge overlapping or adjacent rectangles
func mergeRects(rects []xproto.Rectangle) []xproto.Rectangle {
	if len(rects) == 0 {
		return rects
	}
	// Sort by X, then Y
	sort.Slice(rects, func(i, j int) bool {
		if rects[i].X == rects[j].X {
			return rects[i].Y < rects[j].Y
		}
		return rects[i].X < rects[j].X
	})
	merged := []xproto.Rectangle{rects[0]}
	for _, r := range rects[1:] {
		last := &merged[len(merged)-1]
		// If rectangles overlap or are adjacent, merge them
		if int(r.X) <= int(last.X)+int(last.Width) &&
			int(r.Y) <= int(last.Y)+int(last.Height) {
			// Expand last rectangle to include r
			x1 := min(int(last.X), int(r.X))
			y1 := min(int(last.Y), int(r.Y))
			x2 := max(int(last.X)+int(last.Width), int(r.X)+int(r.Width))
			y2 := max(int(last.Y)+int(last.Height), int(r.Y)+int(r.Height))
			last.X = int16(x1)
			last.Y = int16(y1)
			last.Width = uint16(x2 - x1)
			last.Height = uint16(y2 - y1)
		} else {
			merged = append(merged, r)
		}
	}
	return merged
}

func min(a, b int) int {
	if a < b {
		return a
	} else {
		return b
	}
}
func max(a, b int) int {
	if a > b {
		return a
	} else {
		return b
	}
}

// Frame buffer for latest delta frames
var (
	deltaFrameBuf   []DeltaFrame
	deltaFrameBufMu sync.Mutex
)

// CaptureDeltaJPEG returns a slice of JPEG-encoded merged changed regions
func CaptureDeltaJPEG() ([]DeltaFrame, error) {
	conn, err := getConn()
	if err != nil {
		return nil, err
	}
	rects, err := GetChangedRects()
	if err != nil {
		return nil, err
	}
	if len(rects) == 0 {
		// Return buffered frame if available
		deltaFrameBufMu.Lock()
		defer deltaFrameBufMu.Unlock()
		if len(deltaFrameBuf) > 0 {
			return deltaFrameBuf, nil
		}
		return nil, nil
	}

	// Merge regions for efficiency
	rects = mergeRects(rects)

	setup := xproto.Setup(conn)
	screen := setup.DefaultScreen(conn)
	width := int(screen.WidthInPixels)
	height := int(screen.HeightInPixels)

	reply, err := xproto.GetImage(conn, xproto.ImageFormatZPixmap, xproto.Drawable(screen.Root),
		0, 0, uint16(width), uint16(height), 0xffffffff).Reply()
	if err != nil {
		return nil, err
	}
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	src := reply.Data
	dst := img.Pix
	for i := 0; i+3 < len(src) && i+3 < len(dst); i += 4 {
		dst[i+0] = src[i+2]
		dst[i+1] = src[i+1]
		dst[i+2] = src[i+0]
		dst[i+3] = 255
	}

	var deltas []DeltaFrame
	for _, r := range rects {
		rect := image.Rect(int(r.X), int(r.Y), int(r.X)+int(r.Width), int(r.Y)+int(r.Height))
		cropped := image.NewRGBA(rect)
		draw.Draw(cropped, rect, img, rect.Min, draw.Src)
		buf := new(bytes.Buffer)
		_ = jpeg.Encode(buf, cropped, &jpeg.Options{Quality: 80})
		deltas = append(deltas, DeltaFrame{
			X:      int(r.X),
			Y:      int(r.Y),
			Width:  int(r.Width),
			Height: int(r.Height),
			Data:   buf.Bytes(),
		})
	}
	// Update frame buffer
	deltaFrameBufMu.Lock()
	deltaFrameBuf = deltas
	deltaFrameBufMu.Unlock()
	return deltas, nil
}

// Helper: decode JPEG delta frame to image.Image (for local testing or client-side Go)
func DecodeDeltaJPEG(data []byte) (image.Image, error) {
	return jpeg.Decode(bytes.NewReader(data))
}
