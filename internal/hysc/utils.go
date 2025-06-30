package hysc

import (
	"fmt"
	"image"
	"strconv"
	"strings"

	"github.com/gotk3/gotk3/gdk"
)

func getHandle(address string) (uint64, error) {
	prefix := "0x"

	address = strings.TrimPrefix(address, prefix)

	handle, err := strconv.ParseUint(address, 16, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse address: %v", err)
	}

	return handle, nil
}

func nRGBAtoPixbuf(img *image.NRGBA) (*gdk.Pixbuf, error) {
	width := img.Bounds().Dx()
	height := img.Bounds().Dy()

	pixbuf, err := gdk.PixbufNew(gdk.COLORSPACE_RGB, true, 8, width, height)
	if err != nil {
		return nil, err
	}

	pixels := pixbuf.GetPixels()

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			srcOffset := y*img.Stride + x*4
			dstOffset := (y*width + x) * 4

			copy(
				pixels[dstOffset:dstOffset+4],
				img.Pix[srcOffset:srcOffset+4],
			)
		}
	}

	return pixbuf, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
