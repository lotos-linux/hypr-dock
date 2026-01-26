package ipc

import (
	"strings"
	"sync"
)

type EventListener struct {
	Event   string
	Handler func(string)
	ID      int
	running bool
}

type eventManager struct {
	eventListeners  []*EventListener
	listenerCounter int
	mu              sync.Mutex
}

var (
	eventManagerInstance *eventManager
	once                 sync.Once
)

func getEventManager() *eventManager {
	once.Do(func() {
		eventManagerInstance = &eventManager{
			eventListeners:  make([]*EventListener, 0),
			listenerCounter: 0,
		}
	})
	return eventManagerInstance
}

func AddEventListener(event string, handler func(string), running bool) *EventListener {
	em := getEventManager()
	em.mu.Lock()
	defer em.mu.Unlock()

	listener := &EventListener{
		Event:   event,
		Handler: handler,
		ID:      em.listenerCounter,
		running: running,
	}

	em.eventListeners = append(em.eventListeners, listener)
	em.listenerCounter++
	return listener
}

func DispatchEvent(event string) {
	// log.Printf("IPC Debug: event: %s", event)

	em := getEventManager()
	em.mu.Lock()
	defer em.mu.Unlock()

	for _, listener := range em.eventListeners {
		if strings.Contains(event, listener.Event) && listener.running {
			listener.Handler(event)
		}
	}
}

func (el *EventListener) Run() {
	em := getEventManager()
	em.mu.Lock()
	defer em.mu.Unlock()

	for i := range em.eventListeners {
		if em.eventListeners[i].ID == el.ID {
			em.eventListeners[i].running = true
			break
		}
	}
}

func (el *EventListener) Pause() {
	em := getEventManager()
	em.mu.Lock()
	defer em.mu.Unlock()

	for i := range em.eventListeners {
		if em.eventListeners[i].ID == el.ID {
			em.eventListeners[i].running = false
			break
		}
	}
}

func (el *EventListener) IsRunning() bool {
	em := getEventManager()
	em.mu.Lock()
	defer em.mu.Unlock()

	for i := range em.eventListeners {
		if em.eventListeners[i].ID == el.ID {
			return em.eventListeners[i].running
		}
	}
	return false
}

func (el *EventListener) Remove() {
	em := getEventManager()
	em.mu.Lock()
	defer em.mu.Unlock()

	for i := range em.eventListeners {
		if em.eventListeners[i].ID == el.ID {
			em.eventListeners = append(em.eventListeners[:i], em.eventListeners[i+1:]...)
			break
		}
	}
}
