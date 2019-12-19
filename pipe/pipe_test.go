package pipe

import "testing"

func TestPIPE(t *testing.T) {
	pipe := NewPIPE()
	defer pipe.GetReadPipe().Close()
	defer pipe.GetWritePipe().Close()
}
