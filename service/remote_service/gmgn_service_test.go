package remote_service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQueryTokensByName(t *testing.T) {
	name := "HOKK"
	chain := "Ethereum"

	res, err := QueryTokensByName(name, chain)
	assert.NoError(t, err)
	t.Log(res)
}
