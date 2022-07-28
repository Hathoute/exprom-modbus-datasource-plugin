package helper_test

import (
	"testing"
	"time"
)

func Test(t *testing.T) {
	epoch := 1659028780
	tt := time.Unix(int64(epoch), 0)
	loc := tt.Location()
	t.Log(*loc)
}
