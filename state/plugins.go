package state

/*****************
	State Calls
******************/

// Only called once to set the etcd flag.
func (s *State) SetUseEtcd(setVal bool) {
	s.useEtcd = setVal
}

func (s *State) UsingEtcd() bool {
	return s.useEtcd
}
