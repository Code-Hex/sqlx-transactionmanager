package sqlx

type RollbackErr struct{}

func (r *RollbackErr) Error() string { return "" }

type CommitErr struct{}

func (c *CommitErr) Error() string { return "" }
