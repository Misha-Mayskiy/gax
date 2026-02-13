package sfu

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPeer_Creation(t *testing.T) {
	peer := &Peer{
		ID: "peer-1",
	}
	assert.Equal(t, "peer-1", peer.ID)
}
