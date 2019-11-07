package leader

import (
	"fmt"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages/electionMsgs"
)

func (l *Leader) _execute(m interfaces.IMsg) {
	panic("unused")
	switch m.Type() {
	case constants.SYNC_MSG:
		l.LeaderExecuteSyncMsg(m.(*electionMsgs.SyncMsg))
	case constants.EOM_MSG:
		l.LeaderExecuteEOM(m)
	case constants.DIRECTORY_BLOCK_SIGNATURE_MSG:
		l.LeaderExecuteDBSig(m)
	case constants.DBSTATE_MISSING_MSG:
		m.FollowerExecute(l.State)
	case constants.FACTOID_TRANSACTION_MSG:
		l.LeaderExecute(m)
	case constants.COMMIT_CHAIN_MSG:
		l.LeaderExecuteCommitChain(m)
	case constants.REVEAL_ENTRY_MSG:
		l.LeaderExecuteRevealEntry(m)
	default:
		panic(fmt.Sprintf("leader doesn't handle %v yet ", m.Type()))
	}
	/*
	   // FIXME Apply Messages that need ACK
	   func (m *CommitEntryMsg) LeaderExecute(state interfaces.ILeader) {
	   	state.LeaderExecuteCommitEntry(m)
	   }
	   func (m *ChangeServerKeyMsg) LeaderExecute(state interfaces.ILeader) {
	   	state.LeaderExecute(m)
	   }
	   func (m *RemoveServerMsg) LeaderExecute(state interfaces.ILeader) {
	   	state.LeaderExecute(m)
	   }
	*/
}


/*
// REVIEW: do we need to handle these messages in leader thread
func (m *Ack) LeaderExecute(state interfaces.ILeader) { m.FollowerExecute(state.GetState()) }
func (m *AddServerMsg) ComputeVMIndex(state interfaces.IState) { m.VMIndex = state.ComputeVMIndex(constants.ADMIN_CHAINID) }
func (m *Bounce) LeaderExecute(state interfaces.ILeader) { m.processed = true }
func (m *BounceReply) LeaderExecute(state interfaces.ILeader) { }
func (m *DataResponse) LeaderExecute(state interfaces.ILeader) { m.FollowerExecute(state.GetState()) }
func (m *DBStateMsg) LeaderExecute(state interfaces.ILeader) { m.FollowerExecute(state.GetState()) }
func (m *AddAuditInternal) LeaderExecute(state interfaces.ILeader) { m.FollowerExecute(state.GetState()) }
func (m *AddLeaderInternal) LeaderExecute(state interfaces.ILeader) { m.FollowerExecute(state.GetState()) }
func (m *AuthorityListInternal) LeaderExecute(state interfaces.ILeader) { m.FollowerExecute(state.GetState()) }
func (m *EomSigInternal) LeaderExecute(state interfaces.ILeader) { m.FollowerExecute(state.GetState()) }
func (m *FedVoteLevelMsg) LeaderExecute(state interfaces.ILeader) { m.FollowerExecute(state.GetState()) }
func (m *FedVoteMsg) LeaderExecute(state interfaces.ILeader) { m.FollowerExecute(state.GetState()) }
func (m *FedVoteProposalMsg) LeaderExecute(state interfaces.ILeader) { m.FollowerExecute(state.GetState()) }
func (m *FedVoteVolunteerMsg) LeaderExecute(state interfaces.ILeader) { m.FollowerExecute(state.GetState()) }
func (m *RemoveAuditInternal) LeaderExecute(state interfaces.ILeader) { m.FollowerExecute(state.GetState()) }
func (m *RemoveLeaderInternal) LeaderExecute(state interfaces.ILeader) { m.FollowerExecute(state.GetState()) }
func (m *StartElectionInternal) LeaderExecute(state interfaces.ILeader) { m.FollowerExecute(state.GetState()) }
func (m *SyncMsg) LeaderExecute(l interfaces.ILeader) { l.Enqueue(m) }
func (m *TimeoutInternal) LeaderExecute(state interfaces.ILeader) { m.FollowerExecute(state.GetState()) }
func (m *EOM) LeaderExecute(state interfaces.ILeader) { state.Enqueue(m) }
func (m *Heartbeat) LeaderExecute(state interfaces.ILeader) { m.FollowerExecute(state.GetState()) }
func (m *MissingData) LeaderExecute(state interfaces.ILeader) { m.FollowerExecute(state.GetState()) }
func (m *MissingMsg) LeaderExecute(state interfaces.ILeader) { m.FollowerExecute(state.GetState()) }
func (m *MissingMsgResponse) LeaderExecute(state interfaces.ILeader) { m.FollowerExecute(state.GetState()) }
*/
