package arch

import (
	"testing"
)

func TestServiceMaker(t *testing.T) {
	service := NewSlave("slave", "127.0.0.1:3000", "127.0.0.1:4000")

	sc := Services(service)

	if sc == nil {
		t.Fatalf("service fails to meet services criteria", sc, service)
	}
}
