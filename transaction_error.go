package sqlx

const (
	commitErrMsg  = "Tried to commit but already rollbacked in nested transaction"
	beginTxErrMsg = "Trying to start a transaction in nested state"
)

type NestedCommitErr struct{}

func (n *NestedCommitErr) Error() string {
	return commitErrMsg
}

type NestedBeginTxErr struct{}

func (n *NestedBeginTxErr) Error() string {
	return beginTxErrMsg
}
