package sqlx

type NestedCommitErr struct{}

func (n *NestedCommitErr) Error() string {
	return "Tried to commit but already rollbacked in nested transaction"
}

type NestedBeginTxErr struct{}

func (n *NestedBeginTxErr) Error() string {
	return "Trying to start a transaction in nested state"
}
