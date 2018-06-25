package sqlx

const (
	commitErrMsg  = "Tried to commit but already rollbacked in nested transaction"
	beginTxErrMsg = "Trying to start a transaction in nested state"
)

// NestedCommitErr is an error type to notice that
// commit in nested transaction.
type NestedCommitErr struct{}

func (n *NestedCommitErr) Error() string {
	return commitErrMsg
}
