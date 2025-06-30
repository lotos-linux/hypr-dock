package preview

import (
	"fmt"
	"hypr-dock/internal/item"
	"hypr-dock/internal/settings"
)

type PV struct {
	className string
}

func New() (*PV, error) {
	return &PV{
		className: "",
	}, nil
}

func (pv *PV) Show(item *item.Item, settings settings.Settings) {
	if pv == nil || item == nil {
		fmt.Println("pv is nil:", pv == nil, "|", "item is nil:", item == nil)
		return
	}

	if pv.className == item.ClassName {
		return
	}

	fmt.Println("Show", item.ClassName)

	pv.className = item.ClassName
}

func (pv *PV) Hide(item *item.Item, settings settings.Settings) {

	fmt.Println("Hide", item.ClassName)
}

func (pv *PV) Move(item *item.Item, settings settings.Settings) {
	if pv == nil || item == nil {
		fmt.Println("pv is nil:", pv == nil, "|", "item is nil:", item == nil)
		return
	}

	if pv.className == item.ClassName {
		return
	}

	fmt.Println("Move", item.ClassName)

	pv.className = item.ClassName
}
