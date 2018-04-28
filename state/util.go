package state

import (
	"fmt"

	"time"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/globals"
	"github.com/gonum/stat"
)

type DebugTravelStatistics struct {
	Size          int
	FullPathStats map[byte]*AllPaths

	Name string
}

func NewDebugTravelStatistics(size int, name string) *DebugTravelStatistics {
	n := new(DebugTravelStatistics)
	n.Size = size
	n.FullPathStats = make(map[byte]*AllPaths)
	n.Name = name
	go n.Run()
	return n
}

func (s *DebugTravelStatistics) Run() {
	ticker := time.NewTicker(time.Second * 3)
	for range ticker.C {
		fmt.Println(s.String(s.Name))
	}
}

func (s *DebugTravelStatistics) String(name string) string {
	str := fmt.Sprintf("%s Message Traveling\n", name)

	ft := "%15s: %20s: %s\n"

	for k, v := range s.FullPathStats {
		msgName := constants.MessageName(k)
		str += fmt.Sprintf(ft,
			"Full",
			msgName,
			v.Full.String())
		str += fmt.Sprintf(ft,
			"BroadcatIn",
			msgName,
			v.ToBroadCast.String())
		str += fmt.Sprintf(ft,
			"ToPeerRec",
			msgName,
			v.PeerRec.String())
	}

	return str
}

func (s *DebugTravelStatistics) AddSample(dt globals.DebugTravel) {
	if len(dt.Times) < 1 {
		return
	}
	if _, ok := s.FullPathStats[dt.MsgType]; !ok {
		s.FullPathStats[dt.MsgType] = NewAllPaths(s.Size)
	}
	// time.Since(msg.GetDebugTimestamp().DebugTime).Seconds()
	now := time.Now()

	s.FullPathStats[dt.MsgType].Full.AddTime(now.Sub(dt.DebugTime).Seconds())
	s.FullPathStats[dt.MsgType].ToBroadCast.AddTime(now.Sub(dt.Times[1]).Seconds())
	s.FullPathStats[dt.MsgType].PeerRec.AddTime(now.Sub(dt.Times[2]).Seconds())
}

type AllPaths struct {
	Full        *SinglePath
	ToBroadCast *SinglePath
	PeerRec     *SinglePath
}

func NewAllPaths(size int) *AllPaths {
	s := new(AllPaths)
	s.Full = NewSinglePath(size)
	s.ToBroadCast = NewSinglePath(size)
	s.PeerRec = NewSinglePath(size)
	return s
}

type SinglePath struct {
	Samples   []float64
	MaxSample float64
	MinSample float64

	loops int
	index int
	off   bool
}

func NewSinglePath(size int) *SinglePath {
	s := new(SinglePath)
	s.Samples = make([]float64, size)
	if size == 0 {
		s.off = true
	}
	return s
}

func (s *SinglePath) String() string {
	return fmt.Sprintf("Avg: %5f, Std: %5f, Min: %5f, Max: %5f, I: %3d, L: %2d",
		s.Avg(), s.Stdev(), s.Min(), s.Max(), s.index, s.loops)
}

func (s *SinglePath) AddTime(sec float64) {
	if s.off {
		return
	}
	if sec > 100 {
		return
	}
	if s.MinSample == 0 {
		s.MinSample = sec
	}
	if sec < s.MinSample {
		s.MinSample = sec
	}
	if sec > s.MaxSample {
		s.MaxSample = sec
	}
	s.Samples[s.index] = sec
	s.index++
	if s.index == len(s.Samples) {
		s.index = 0
		s.loops++
	}
}

func (s *SinglePath) slice() []float64 {
	if s.loops > 0 {
		return s.Samples
	}
	return s.Samples[:s.index]
}

func (s *SinglePath) Avg() float64 {
	return stat.Mean(s.slice(), nil)
}

func (s *SinglePath) Stdev() float64 {
	return stat.StdDev(s.slice(), nil)
}

func (s *SinglePath) Max() float64 {
	return s.MaxSample
}

func (s *SinglePath) Min() float64 {
	return s.MinSample
}
