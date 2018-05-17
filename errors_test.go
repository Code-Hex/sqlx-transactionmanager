package sqlx

import (
	"testing"
)

func TestErrors(t *testing.T) {
	txer := new(NestedBeginTxErr)
	if txer.Error() != beginTxErrMsg {
		t.Fatal("Something error")
	}

	cterr := new(NestedCommitErr)
	if cterr.Error() != commitErrMsg {
		t.Fatal("Something error")
	}
}
