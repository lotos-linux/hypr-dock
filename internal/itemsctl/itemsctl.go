package itemsctl

import (
	"errors"
	"hypr-dock/internal/item"
	"hypr-dock/pkg/ipc"
)

type List struct {
	list map[string]*item.Item
}

func New() *List {
	return &List{
		list: make(map[string]*item.Item),
	}
}

func (l *List) GetMap() map[string]*item.Item {
	return l.list
}

func (l *List) Get(className string) *item.Item {
	return l.list[className]
}

func (l *List) Add(className string, item *item.Item) {
	l.list[className] = item
}

func (l *List) Remove(className string) {
	delete(l.list, className)
}

func (l *List) Len() int {
	return len(l.list)
}

func (l *List) SearchWindow(address string) (*item.Item, *ipc.Client, error) {
	for _, item := range l.list {
		win, exist := item.Windows[address]
		if exist {
			return item, win, nil
		}
	}

	err := errors.New("Window not found: " + address)
	return nil, nil, err
}
