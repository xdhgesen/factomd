package state

// BalanceMap enables CRUD of balances in a threadsafe manner.
type BalanceMap struct {
	balances map[[32]byte]int64

	queries chan *BalanceQuery

	quit chan struct{}
}

type BalanceQuery struct {
}

func NewBalanceQuery() *BalanceQuery {
	q := new(BalanceQuery)
	return q
}

func NewBalanceMap() *BalanceMap {
	b := new(BalanceMap)

	return b
}

func (bm *BalanceMap) Close() {
	close(bm.quit)
}

// Closed returns true if the BalanceMap is closed and no longer
// servicing returns
func (bm *BalanceMap) Closed() bool {
	select {
	case _, open := <-bm.quit:
		return !open
	default:
		return false
	}
}

func (bm *BalanceMap) Respond(q *BalanceQuery) {
	//bal := bm.balances[q.Address]
	//q.ret <- bal
}

// Serve will enable the serving
func (bm *BalanceMap) Serve() {
	for {
		select {
		case <-bm.quit:
		case q := <-bm.queries:
			bm.Respond(q)
		}
	}
}
