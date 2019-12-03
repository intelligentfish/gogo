package block_chan

import (
	"testing"
)

func TestBlockChan(t *testing.T) {
	blockChan := NewBlockChan()
	ok := blockChan.AddBlock(NewBlockChanNode(nil)).
		AddBlock(NewBlockChanNode(nil)).
		Valid()
	t.Log(ok)
}
