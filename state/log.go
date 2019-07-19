package state

import (
	"github.com/FactomProject/factomd/common/globals"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
)

// StateLogger is the state logging struct that can be passed around to give access to logging
// without having to pass around all of state.
type StateLogger struct {
	// The names are underscore so when using an IDE, the 's.l' does not pop these up.
	// We shouldn't be accessing them, but this struct is in the same package as State,
	// so we can't stop anyone. We can make it awkward though.
	_logFactomNodeName string
	_logHeight         int
	_logMinute         int
}

func NewStateLogger(s *State) *StateLogger {
	l := new(StateLogger)
	l._logFactomNodeName = s.GetFactomNodeName()

	return l
}

func (s *StateLogger) LogPrintf(logName string, format string, more ...interface{}) {
	if s.DebugExec() {
		if s == nil {
			messages.StateLogPrintf("unknown", 0, 0, logName, format, more...)
		} else {
			messages.StateLogPrintf(s._logFactomNodeName, int(s._logHeight), int(s._logMinute), logName, format, more...)
		}
	}
}

func (s *StateLogger) LogMessage(logName string, comment string, msg interfaces.IMsg) {
	if s.DebugExec() {
		if s == nil {
			messages.StateLogMessage("unknown", 0, 0, logName, comment, msg)
		} else {
			messages.StateLogMessage(s._logFactomNodeName, int(s._logHeight), int(s._logMinute), logName, comment, msg)
		}
	}
}

func (s *StateLogger) UpdateLogHeightAndMinute(height, minute int) {
	s._logHeight = height
	s._logMinute = minute
}

func (s *StateLogger) DebugExec() bool {
	return globals.Params.DebugLogRegEx != ""
}

func (s *StateLogger) CheckFileName(name string) bool {
	return messages.CheckFileName(name)
}
