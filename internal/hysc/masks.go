package hysc

import (
	"math"

	"github.com/gotk3/gotk3/gdk"
)

func ApplyVector(p *gdk.Pixbuf, pos Cord, size Size, mask func(Cord) float64) {
	if p.GetNChannels() != 4 {
		return
	}

	pixels := p.GetPixels()
	stride := p.GetRowstride()
	width, height := p.GetWidth(), p.GetHeight()

	x0 := max(0, pos.X)
	y0 := max(0, pos.Y)
	x1 := min(pos.X+size.W, width)
	y1 := min(pos.Y+size.H, height)

	for y := y0; y < y1; y++ {
		rowOffset := y * stride
		relY := y - pos.Y

		for x := x0; x < x1; x++ {
			alpha := mask(Cord{x - pos.X, relY})
			offset := rowOffset + x*4 + 3
			pixels[offset] = uint8(alpha * 255)
		}
	}
}

func radiusmask(pixel Cord, center Cord, R float64) float64 {
	dx := float64(pixel.X - center.X)
	dy := float64(pixel.Y - center.Y)
	distance := math.Sqrt(dx*dx + dy*dy)

	if distance <= R-1.0 {
		return 1.0
	}
	if distance >= R+1.0 {
		return 0.0
	}

	t := (distance - (R - 1.0)) / 2.0
	return 1.0 - t*t
}
