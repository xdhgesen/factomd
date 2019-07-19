package state

import (
	"fmt"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

// TimestampModules contains the "time" for a factomd node.
//	By keeping it in a struct embedded in factomd, we can pass around
//	the module without passing around all of state
type TimestampModules struct {
	IsReplaying     bool
	ReplayTimestamp interfaces.Timestamp
	TimeOffset      interfaces.Timestamp
	MaxTimeOffset   interfaces.Timestamp
}

// Returns a millisecond timestamp
func (s *TimestampModules) GetTimestamp() interfaces.Timestamp {
	if s.IsReplaying == true {
		fmt.Println("^^^^^^^^ IsReplying is true")
		return s.ReplayTimestamp
	}
	return primitives.NewTimestampNow()
}

func (s *TimestampModules) SetIsReplaying() {
	s.IsReplaying = true
}

func (s *TimestampModules) SetIsDoneReplaying() {
	s.IsReplaying = false
	s.ReplayTimestamp = nil
}
