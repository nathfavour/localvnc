package screen

import (
	"bytes"
	"github.com/kbinani/screenshot"
	"image/png"
)

func CapturePNG() ([]byte, error) {
	img, err := screenshot.CaptureDisplay(0)
	if err != nil {
		return nil, err
	}
	buf := new(bytes.Buffer)
	err = png.Encode(buf, img)
	return buf.Bytes(), err
}
