package globals

import "time"

type DebugTravel struct {
	DebugTime time.Time
	Path      string
	Paths     []string
	Times     []time.Time
	MsgType   byte
}

var x time.Time

func (d *DebugTravel) Start(me string, t byte) {
	d.Path = me
	d.DebugTime = time.Now()
	d.MsgType = t
	d.Times = []time.Time{}
	d.Paths = []string{}
}

func (d *DebugTravel) Touch(me string) {
	d.Path += " -> " + me
	d.Paths = append(d.Paths, me)
	nt := time.Now()
	d.Times = append(d.Times, nt)

	if nt.Before(d.DebugTime) {
		panic("impoissible")
	}
}
