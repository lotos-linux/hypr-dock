package popup

import (
	"errors"

	"github.com/dlasky/gotk3-layershell/layershell"
	"github.com/gotk3/gotk3/gtk"
)

type Popup struct {
	x, y           int
	acitve         bool
	win            *gtk.Window
	content        gtk.IWidget
	xstart, ystart string

	winCallBack func(*gtk.Window) error
}

func New() *Popup {
	return &Popup{
		x:           0,
		y:           0,
		acitve:      false,
		win:         nil,
		content:     nil,
		winCallBack: nil,
	}
}

func (p *Popup) Open(x, y int, xstart, ystart string) error {
	p.Close()

	win, err := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	if err != nil {
		return err
	}

	p.win = win
	if p.winCallBack != nil {
		p.winCallBack(win)
	}
	initLayerShell(win)

	p.x = x
	p.y = y
	p.xstart = xstart
	p.ystart = ystart

	if p.content != nil {
		p.win.Add(p.content)
	}
	p.setCord()
	p.win.ShowAll()
	p.acitve = true

	return nil
}

func (p *Popup) Close() {
	p.acitve = false
	if p.win != nil {
		p.win.Destroy()
		p.win = nil
	}
}

func (p *Popup) Set(content gtk.IWidget) error {
	if content == nil {
		return errors.New("content is nil")
	}

	p.content = content
	if !p.acitve {
		return nil
	}

	child, err := p.win.GetChild()
	if err == nil && child != nil {
		child.ToWidget().Destroy()
	}

	p.win.Add(content)
	p.win.ShowAll()

	return nil
}

func (p *Popup) Move(x, y int) {
	p.x = x
	p.y = y

	if p.acitve {
		p.setCord()
	}
}

func (p *Popup) Shift(dx, dy int) {
	p.x = p.x + dx
	p.y = p.y + dy

	if p.acitve {
		p.setCord()
	}
}

func (p *Popup) SetWinCallBack(callback func(*gtk.Window) error) {
	p.winCallBack = callback
}

func (p *Popup) setCord() {
	if p.win == nil {
		return
	}

	xstarts := map[string]layershell.LayerShellEdgeFlags{
		"left":  layershell.LAYER_SHELL_EDGE_LEFT,
		"right": layershell.LAYER_SHELL_EDGE_RIGHT,
	}

	ystarts := map[string]layershell.LayerShellEdgeFlags{
		"top":    layershell.LAYER_SHELL_EDGE_TOP,
		"bottom": layershell.LAYER_SHELL_EDGE_BOTTOM,
	}

	layershell.SetAnchor(p.win, xstarts[p.xstart], true)
	layershell.SetAnchor(p.win, ystarts[p.ystart], true)
	layershell.SetMargin(p.win, xstarts[p.xstart], p.x)
	layershell.SetMargin(p.win, ystarts[p.ystart], p.y)
}

func initLayerShell(win *gtk.Window) {
	layershell.InitForWindow(win)
	layershell.SetNamespace(win, "dock-popup")
	layershell.SetLayer(win, layershell.LAYER_SHELL_LAYER_TOP)
}
