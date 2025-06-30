package wl

import (
	"fmt"
	"image"
	"image/color"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/pdf/go-wayland/client"
	"golang.org/x/sys/unix"
)

type App struct {
	display  *client.Display
	registry *client.Registry
	shm      *client.Shm
	tl       *HyprlandToplevelExportManagerV1
	log      hclog.Logger
}

type shmPool struct {
	*client.ShmPool
	fd   int
	data []byte
}

type FrameStream struct {
	Frames chan *image.NRGBA
	stop   chan struct{}
}

func NewApp(log hclog.Logger) (*App, error) {
	display, err := client.Connect(``)
	if err != nil {
		return nil, err
	}

	registry, err := display.GetRegistry()
	if err != nil {
		return nil, err
	}

	app := &App{
		display:  display,
		registry: registry,
		log:      log.Named(`wl`),
	}
	display.SetErrorHandler(app.handleDisplayError)
	registry.SetGlobalHandler(app.handleRegistryGlobal)

	// init registry
	if err := app.roundTrip(); err != nil {
		return nil, err
	}

	// get events
	if err := app.roundTrip(); err != nil {
		return nil, err
	}

	return app, nil
}

func (a *App) StartStream(handle uint64, fps int, bufferSize int) (*FrameStream, error) {
	stream := &FrameStream{
		Frames: make(chan *image.NRGBA, bufferSize),
		stop:   make(chan struct{}),
	}

	go func() {
		ticker := time.NewTicker(time.Second / time.Duration(fps))
		defer ticker.Stop()

		for {
			select {
			case <-stream.stop:
				close(stream.Frames)
				return
			case <-ticker.C:
				frame, err := a.CaptureFrame(handle)
				if err != nil {
					a.log.Trace("Capture error: %v", err)
					continue
				}

				select {
				case stream.Frames <- frame:
				default:
					a.log.Trace("Buffer full")
				}
			}
		}
	}()

	return stream, nil
}

func (s *FrameStream) Stop() {
	close(s.stop)
}

func (a *App) CaptureFrame(handle uint64) (*image.NRGBA, error) {
	if a.tl == nil {
		return nil, fmt.Errorf(`toplevel export manager not available`)
	}

	frame, err := a.tl.CaptureToplevel(0, uint32(handle))
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := frame.Destroy(); err != nil {
			a.log.Error(`failed destroying frame`, `err`, err)
		}
	}()

	formats := make([]HyprlandToplevelExportFrameV1BufferEvent, 0)
	done := make(chan struct{})
	ready := make(chan struct{})
	failed := make(chan error, 1)
	frame.SetBufferHandler(func(evt HyprlandToplevelExportFrameV1BufferEvent) {
		formats = append(formats, evt)
	})
	frame.SetBufferDoneHandler(func(evt HyprlandToplevelExportFrameV1BufferDoneEvent) {
		close(done)
	})
	frame.SetReadyHandler(func(evt HyprlandToplevelExportFrameV1ReadyEvent) {
		close(ready)
	})
	frame.SetFailedHandler(func(evt HyprlandToplevelExportFrameV1FailedEvent) {
		failed <- fmt.Errorf(`frame failed`)
	})

	if err := a.roundTrip(); err != nil {
		return nil, err
	}

	select {
	case <-done:
	case err := <-failed:
		return nil, err
	}

	if len(formats) == 0 {
		return nil, fmt.Errorf(`no buffer formats`)
	}

	var selected *HyprlandToplevelExportFrameV1BufferEvent
OUTER:
	for _, format := range formats {
		switch client.ShmFormat(format.Format) {
		case client.ShmFormatArgb8888:
			selected = &format
			break OUTER
		case client.ShmFormatXrgb8888:
			selected = &format
			break OUTER
		}
	}

	if selected == nil {
		return nil, fmt.Errorf(`no suitable buffer format`)
	}

	pool, err := a.createShmPool(int32(selected.Height * selected.Stride))
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := pool.Close(); err != nil {
			a.log.Error(`failed closing SHM pool`, `err`, err)
		}
	}()

	buf, err := pool.CreateBuffer(0, int32(selected.Width), int32(selected.Height), int32(selected.Stride), selected.Format)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := buf.Destroy(); err != nil {
			a.log.Error(`failed destroying buffer`, `err`, err)
		}
	}()

	if err := frame.Copy(buf, 1); err != nil {
		return nil, err
	}

	if err := a.roundTrip(); err != nil {
		return nil, err
	}

	select {
	case <-ready:
	case err := <-failed:
		return nil, err
	}

	data := pool.Data()
	img := image.NewNRGBA(image.Rect(0, 0, int(selected.Width), int(selected.Height)))
	if len(img.Pix) < int(selected.Height)*int(selected.Stride) {
		return nil, fmt.Errorf(`image buffer too small`)
	}
	for y := range int(selected.Height) {
		for x := range int(selected.Width) {
			pix := data[y*int(selected.Stride)+(x*4) : y*int(selected.Stride)+(x*4)+4]
			col := color.NRGBA{}
			switch client.ShmFormat(selected.Format) {
			case client.ShmFormatArgb8888:
				col.A = pix[3]
				col.R = pix[2]
				col.G = pix[1]
				col.B = pix[0]
			case client.ShmFormatXrgb8888:
				col.A = 0xff
				col.R = pix[2]
				col.G = pix[1]
				col.B = pix[0]
			}
			img.SetNRGBA(x, y, col)
		}
	}

	return img, nil
}

func (a *App) Close() error {
	if a.tl != nil {
		if err := a.tl.Destroy(); err != nil {
			return err
		}
	}
	if a.shm != nil {
		if err := a.shm.Release(); err != nil {
			return nil
		}
	}
	if a.registry != nil {
		if err := a.registry.Destroy(); err != nil {
			return err
		}
	}
	if a.display != nil {
		if err := a.display.Destroy(); err != nil {
			return err
		}
	}
	return nil
}

func (p *shmPool) Data() []byte {
	return p.data
}

func (p *shmPool) Close() error {
	if err := unix.Munmap(p.data); err != nil {
		return err
	}
	if err := p.Destroy(); err != nil {
		return err
	}
	if err := unix.Close(p.fd); err != nil {
		return err
	}
	return nil
}

func (a *App) handleDisplayError(evt client.DisplayErrorEvent) {
	a.log.Trace("Display error occurred", "error", evt)

	err := a.reconnect()
	if err != nil {
		a.log.Trace("Reconnection failed", "error", err)
	}

	a.log.Trace("Successfully reconnected to Wayland display")
}

func (a *App) handleShmFormat(evt client.ShmFormatEvent) {
	a.log.Trace(`reported available SHM format`, `format`, client.ShmFormat(evt.Format))
}

func (a *App) handleRegistryGlobal(evt client.RegistryGlobalEvent) {
	a.log.Trace(`global object`, `name`, evt.Name, `interface`, evt.Interface, `version`, evt.Version)

	switch evt.Interface {
	case `wl_shm`:
		shm := client.NewShm(a.display.Context())
		if err := a.registry.Bind(evt.Name, evt.Interface, evt.Version, shm); err != nil {
			a.log.Error(`failed binding SHM`, `err`, err)
			return
		}
		shm.SetFormatHandler(a.handleShmFormat)
		a.shm = shm
	case `hyprland_toplevel_export_manager_v1`:
		tl := NewHyprlandToplevelExportManagerV1(a.display.Context())
		if err := a.registry.Bind(evt.Name, evt.Interface, evt.Version, tl); err != nil {
			a.log.Error(`failed binding toplevel export manager`, `err`, err)
			return
		}
		a.tl = tl
	}
}

func (a *App) reconnect() error {
	a.Close()

	display, err := client.Connect("")
	if err != nil {
		return fmt.Errorf("failed to reconnect: %w", err)
	}

	registry, err := display.GetRegistry()
	if err != nil {
		return fmt.Errorf("failed to get registry: %w", err)
	}

	a.display = display
	a.registry = registry

	display.SetErrorHandler(a.handleDisplayError)
	registry.SetGlobalHandler(a.handleRegistryGlobal)

	if err := a.roundTrip(); err != nil {
		return err
	}

	return a.roundTrip()
}

func (a *App) createShmPool(size int32) (*shmPool, error) {
	fd, err := unix.MemfdSecret(0)
	if err != nil {
		return nil, fmt.Errorf(`failed creating memfd: %w`, err)
	}
	if err := unix.Ftruncate(fd, int64(size)); err != nil {
		return nil, fmt.Errorf(`failed truncating memfd: %w`, err)
	}

	data, err := unix.Mmap(fd, 0, int(size), unix.PROT_READ|unix.PROT_WRITE, unix.MAP_SHARED)
	if err != nil {
		return nil, fmt.Errorf(`failed mmapping memfd: %w`, err)
	}

	pool, err := a.shm.CreatePool(fd, int32(size))
	if err != nil {
		return nil, fmt.Errorf(`failed creating SHM pool: %w`, err)
	}

	return &shmPool{
		ShmPool: pool,
		fd:      fd,
		data:    data,
	}, nil
}

func (a *App) roundTrip() error {
	cb, err := a.display.Sync()
	if err != nil {
		return err
	}
	defer func() {
		if err := cb.Destroy(); err != nil {
			a.log.Error(`failed destroying callback`, `err`, err)
		}
	}()

	done := make(chan struct{})
	cb.SetDoneHandler(func(_ client.CallbackDoneEvent) {
		close(done)
	})

	for {
		select {
		case <-done:
			return nil
		default:
			if err := a.display.Context().Dispatch(); err != nil {
				a.log.Trace(`dispatch error`, `err`, err)
			}
		}
	}
}
