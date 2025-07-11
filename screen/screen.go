package screen

import (
	"bytes"
	"image"
	"image/jpeg"
	"image/png"
	"math/rand"
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
