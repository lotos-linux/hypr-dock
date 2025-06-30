package hysc

import (
	"fmt"
	"hypr-dock/pkg/wl"
	"log"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/hashicorp/go-hclog"
)

type Stream struct {
	address string
	handle  uint64

	size       *Size
	scaleMode  *ScaleMode
	interpType gdk.InterpType

	effects map[string]func(p *gdk.Pixbuf) error
	masks   map[string]func(p *gdk.Pixbuf) error

	readyHandler func(*Size)
	frameHandler func(*Size)
	errorHandler func(error)

	*gtk.Image
}

type ScaleMode struct {
	scaleW int
	scaleH int
	scaleF float64
}

type Size struct {
	W, H int
}

type Cord struct {
	X, Y int
}

func StreamAndStart(address string, fps int) (*Stream, error) {
	s, err := StreamNew(address)
	if err != nil {
		return nil, err
	}

	err = s.Start(fps)
	if err != nil {
		return nil, err
	}

	return s, nil
}

func StreamNew(address string) (*Stream, error) {
	handle, err := getHandle(address)
	if err != nil {
		return nil, err
	}

	widget, err := gtk.ImageNew()
	if err != nil {
		return nil, err
	}

	return &Stream{
		address: address,
		handle:  handle,

		size:       nil,
		scaleMode:  nil,
		interpType: gdk.INTERP_BILINEAR,

		effects: make(map[string]func(p *gdk.Pixbuf) error, 0),
		masks:   make(map[string]func(p *gdk.Pixbuf) error, 0),

		readyHandler: nil,
		frameHandler: nil,
		errorHandler: func(err error) {
			log.Println(err)
		},

		Image: widget,
	}, nil
}

func (s *Stream) Start(fps int, buferSize ...int) error {
	app, err := wl.NewApp(hclog.Default())
	if err != nil {
		return fmt.Errorf("failed to wayland connection: %v", err)
	}

	var bufSize int
	if len(buferSize) == 0 {
		bufSize = 5
	} else {
		bufSize = buferSize[0]
	}

	stream, err := app.StartStream(s.handle, fps, bufSize)
	if err != nil {
		app.Close()
		return fmt.Errorf("failed to start stream: %v", err)
	}

	go func() {
		for frame := range stream.Frames {
			glib.IdleAdd(func() {
				pixbuf, err := nRGBAtoPixbuf(frame)
				if err != nil {
					s.errorHandler(fmt.Errorf("failed to convert pixbuf: %v", err))
					return
				}

				if s.scaleMode != nil {
					pixbuf, err = s.scale(pixbuf)
					if err != nil {
						s.errorHandler(fmt.Errorf("failed to scale pixbuf: %v", err))
						return
					}
				}

				for id, effect := range s.effects {
					err = effect(pixbuf)
					if err != nil {
						s.errorHandler(fmt.Errorf("'%s' effect error: %v", id, err))
						return
					}
				}

				for id, mask := range s.masks {
					err = mask(pixbuf)
					if err != nil {
						s.errorHandler(fmt.Errorf("'%s' mask error: %v", id, err))
						return
					}
				}

				size := &Size{
					W: pixbuf.GetWidth(),
					H: pixbuf.GetHeight(),
				}

				s.SetFromPixbuf(pixbuf)

				if s.size == nil && s.readyHandler != nil {
					s.readyHandler(size)
				}

				if s.frameHandler != nil {
					s.frameHandler(size)
				}

				s.size = size
			})
		}
	}()

	s.Connect("destroy", func() {
		stream.Stop()
		app.Close()
		s = nil
	})

	return nil
}

func (s *Stream) CaptureFrame() error {
	app, err := wl.NewApp(hclog.Default())
	if err != nil {
		return fmt.Errorf("failed to wayland connection: %v", err)
	}
	defer app.Close()

	frame, err := app.CaptureFrame(s.handle)
	if err != nil {
		app.Close()
		return fmt.Errorf("failed to capture frame: %v", err)
	}

	pixbuf, err := nRGBAtoPixbuf(frame)
	if err != nil {
		return fmt.Errorf("failed to convert pixbuf: %v", err)
	}

	if s.scaleMode != nil {
		pixbuf, err = s.scale(pixbuf)
		if err != nil {
			return fmt.Errorf("failed to scale pixbuf: %v", err)
		}
	}

	for id, effect := range s.effects {
		err = effect(pixbuf)
		if err != nil {
			return fmt.Errorf("'%s' effect error: %v", id, err)
		}
	}

	for id, mask := range s.masks {
		err = mask(pixbuf)
		if err != nil {
			return fmt.Errorf("'%s' mask error: %v", id, err)
		}
	}

	size := &Size{
		W: pixbuf.GetWidth(),
		H: pixbuf.GetHeight(),
	}

	s.SetFromPixbuf(pixbuf)

	if s.readyHandler != nil {
		s.readyHandler(size)
	}

	s.size = size

	return nil
}

func (s *Stream) SetWScale(width int) {
	s.scaleMode = &ScaleMode{scaleW: width}
}

func (s *Stream) SetHScale(height int) {
	s.scaleMode = &ScaleMode{scaleH: height}
}

func (s *Stream) SetFScale(factor float64) {
	s.scaleMode = &ScaleMode{scaleF: factor}
}

func (s *Stream) SetFixedSize(width int, height int) {
	s.scaleMode = &ScaleMode{
		scaleW: width,
		scaleH: height,
	}
}

func (s *Stream) ResetSize() {
	s.scaleMode = nil
}

func (s *Stream) AddCustomEffect(id string, effectHandler func(p *gdk.Pixbuf) error) {
	if effectHandler == nil {
		return
	}

	s.effects[id] = effectHandler
}

func (s *Stream) RemoveCustomEffect(id string) bool {
	_, exist := s.effects[id]
	if exist {
		delete(s.effects, id)
	}

	return exist
}

func (s *Stream) AddCustomMask(id string, maskHandler func(p *gdk.Pixbuf) error) {
	if maskHandler == nil {
		return
	}

	s.masks[id] = maskHandler
}

func (s *Stream) RemoveCustomMask(id string) bool {
	_, exist := s.masks[id]
	if exist {
		delete(s.masks, id)
	}

	return exist
}

func (s *Stream) SetBorderRadius(radius int) {
	id := "system-border-radius:hysc00"

	if radius <= 0 {
		s.RemoveCustomMask(id)
		return
	}

	s.AddCustomMask(id, func(p *gdk.Pixbuf) error {
		w, h := p.GetWidth(), p.GetHeight()
		size := Size{radius, radius}

		ApplyVector(p, Cord{0, 0}, size, func(pixel Cord) float64 {
			return radiusmask(pixel, Cord{radius, radius}, float64(radius))
		})
		ApplyVector(p, Cord{w - radius, 0}, size, func(pixel Cord) float64 {
			return radiusmask(pixel, Cord{0, radius}, float64(radius))
		})
		ApplyVector(p, Cord{0, h - radius}, size, func(pixel Cord) float64 {
			return radiusmask(pixel, Cord{radius, 0}, float64(radius))
		})
		ApplyVector(p, Cord{w - radius, h - radius}, size, func(pixel Cord) float64 {
			return radiusmask(pixel, Cord{0, 0}, float64(radius))
		})
		return nil
	})
}

func (s *Stream) SetInterpType(interpType gdk.InterpType) {
	s.interpType = interpType
}

func (s *Stream) OnReady(handler func(*Size)) {
	s.readyHandler = handler
}

func (s *Stream) OnFrame(handler func(*Size)) {
	s.frameHandler = handler
}

func (s *Stream) OnError(handler func(error)) {
	s.errorHandler = handler
}

func (s *Stream) scale(pixbuf *gdk.Pixbuf) (*gdk.Pixbuf, error) {
	origWidth := pixbuf.GetWidth()
	origHeight := pixbuf.GetHeight()

	// fixed dimensions
	if s.scaleMode.scaleH > 0 && s.scaleMode.scaleW > 0 {
		return pixbuf.ScaleSimple(
			s.scaleMode.scaleW,
			s.scaleMode.scaleH,
			s.interpType,
		)
	}

	// fixed height (width proportional)
	if s.scaleMode.scaleH > 0 {
		newHeight := s.scaleMode.scaleH
		ratio := float64(origWidth) / float64(origHeight)
		newWidth := int(float64(newHeight) * ratio)

		return pixbuf.ScaleSimple(
			newWidth,
			newHeight,
			s.interpType,
		)
	}

	// fixed width (height proportional)
	if s.scaleMode.scaleW > 0 {
		newWidth := s.scaleMode.scaleW
		ratio := float64(origHeight) / float64(origWidth)
		newHeight := int(float64(newWidth) * ratio)

		return pixbuf.ScaleSimple(
			newWidth,
			newHeight,
			s.interpType,
		)
	}

	// scaling by factor
	if s.scaleMode.scaleF > 0 {
		factor := s.scaleMode.scaleF
		newWidth := int(float64(origWidth) * factor)
		newHeight := int(float64(origHeight) * factor)

		return pixbuf.ScaleSimple(
			newWidth,
			newHeight,
			s.interpType,
		)
	}

	// no scaling
	return pixbuf, nil
}
