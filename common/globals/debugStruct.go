package globals

import "time"

type DebugTravel struct {
	DebugTime time.Time
	Path      string
}

func (d *DebugTravel) Start(me string) {
	d.Path = me
	d.DebugTime = time.Now()
}

func (d *DebugTravel) Touch(me string) {
	d.Path += " -> " + me
}
