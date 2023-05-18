package services

import (
	"bytes"
	"errors"

	"github.com/YZhenY/tendermint/libs/common"
)

func (app *ABCIApp) ValidateAndUpdateAndTagBFTTx(tx []byte) (bool, *[]common.KVPair, error) {
	// Validate prefix tx, bft maybe fake
	var tags []common.KVPair
	if bytes.Compare([]byte("mug00"), tx[:len([]byte("mug00"))]) != 0 {
		return false, &tags, errors.New("Tx signature is not mug00")
	}
	return false, &tags, nil
}
