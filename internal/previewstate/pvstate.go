package pvstate

import (
	"hypr-dock/internal/pkg/timer"
	"hypr-dock/internal/preview"
	"log"
)

type PVState struct {
	pv     *preview.PV
	active bool

	showTimer *timer.Timer
	hideTimer *timer.Timer
	moveTimer *timer.Timer

	currentClassName string
}

func New() *PVState {
	pv, err := preview.New()
	if err != nil {
		log.Println(err)
	}

	if pv == nil {
		log.Println("pv is nil")
	}

	return &PVState{
		currentClassName: "90348d332fvecs324csd4",
		showTimer:        timer.New(),
		hideTimer:        timer.New(),
		moveTimer:        timer.New(),
		pv:               pv,
	}
}

func (s *PVState) HasClassChanged(className string) bool {
	return s.currentClassName != className
}

func (s *PVState) SetCurrentClass(ckassName string) {
	s.currentClassName = ckassName
}

func (s *PVState) GetShowTimer() *timer.Timer {
	return s.showTimer
}

func (s *PVState) GetHideTimer() *timer.Timer {
	return s.hideTimer
}

func (s *PVState) GetMoveTimer() *timer.Timer {
	return s.moveTimer
}

func (s *PVState) SetPV(pv *preview.PV) {
	s.pv = pv
}

func (s *PVState) GetPV() *preview.PV {
	return s.pv
}

func (s *PVState) GetActive() bool {
	return s.active
}

func (s *PVState) SetActive(active bool) {
	s.active = active
}
