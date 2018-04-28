package state

import (
	"fmt"

	"github.com/gonum/stat"
)

type Statistic struct {
	Samples   []float64
	MaxSample float64
	MinSample float64

	loops int
	index int
	off   bool
}

func NewStatistic(size int) *Statistic {
	s := new(Statistic)
	s.Samples = make([]float64, size)
	if size == 0 {
		s.off = true
	}
	return s
}

func (s *Statistic) String() string {
	return fmt.Sprintf("Avg: %f, Std: %f, Min: %f, Max: %f, I: %d, L: %d",
		s.Avg(), s.Stdev(), s.Min(), s.Max(), s.index, s.loops)
}

func (s *Statistic) AddTime(sec float64) {
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

func (s *Statistic) slice() []float64 {
	if s.loops > 0 {
		return s.Samples
	}
	return s.Samples[:s.index]
}

func (s *Statistic) Avg() float64 {
	return stat.Mean(s.slice(), nil)
}

func (s *Statistic) Stdev() float64 {
	return stat.StdDev(s.slice(), nil)
}

func (s *Statistic) Max() float64 {
	return s.MaxSample
}

func (s *Statistic) Min() float64 {
	return s.MinSample
}
