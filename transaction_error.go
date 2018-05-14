package sqlx

type NestedRollbackErr struct{}

func (r *NestedRollbackErr) Error() string {
	return "Tried to commit but already rollbacked in nested transaction"
}
