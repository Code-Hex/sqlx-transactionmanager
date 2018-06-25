package sqlx

import (
	"testing"
)

func TestErrors(t *testing.T) {
	cterr := new(NestedCommitErr)
	if cterr.Error() != commitErrMsg {
		t.Fatal("Something error")
	}
}
