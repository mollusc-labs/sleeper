package sleeper_test

import (
	"testing"

	"github.com/mollusc-labs/sleeper"
)

func TestNew(t *testing.T) {
	a := sleeper.NewAuth("foo", "bar")
	c := sleeper.NewConfig("http", 5984, 5000, "127.0.0.1")
	_, err := sleeper.New("posts", c, a)

	if err != nil {
		t.Error("Error should be nil")
	}
}
