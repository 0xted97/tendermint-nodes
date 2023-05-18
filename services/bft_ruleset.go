package services

import (
	"encoding/json"
	"errors"

	"github.com/YZhenY/tendermint/libs/common"
	"github.com/sirupsen/logrus"
)

func (app *ABCIApp) ValidateAndUpdateAndTagBFTTx(tx []byte, msgType byte) (bool, *[]common.KVPair, error) {
	// Validate prefix tx, bft maybe fake
	var tags []common.KVPair
	switch msgType {
	case byte(1): // AssignmentBFTTx
		logrus.Debug("Assignmentbfttx happening")
		if app.state.LastUnassignedIndex >= app.state.LastCreatedIndex {
			return false, &tags, errors.New("Last assigned index is exceeding last created index")
		}

		var parsedTx AssignmentBFTTx
		err := json.Unmarshal(tx, &parsedTx)
		if err != nil {
			logrus.WithError(err).Errorln("Unmarshal AssignmentBFTTx failed")
			return false, &tags, err
		}
		verifierID := parsedTx.VerifierID
		verifier := parsedTx.Verifier
		index := app.getKeyIndex(verifierID, verifier)
		if index >= 0 {
			return false, &tags, errors.New("Key already assigned to verifier")
		}
		app.state.NewKeyAssignments[app.state.LastUnassignedIndex].Verifiers[verifier] = verifierID
		app.state.LastUnassignedIndex = app.state.LastUnassignedIndex + 1

	}
	return true, &tags, nil
}

// TODO: Pre check before DeliverTx
func (app *ABCIApp) validateTx(tx []byte, msgType byte) (bool, error) {
	switch msgType {
	case byte(1): // AssignmentBFTTx
		logrus.Debug("Assignmentbfttx happening")
		// Check available key
		if app.state.LastUnassignedIndex >= app.state.LastCreatedIndex {
			return false, errors.New("Last assigned index is exceeding last created index")
		}
		var parsedTx AssignmentBFTTx
		err := json.Unmarshal(tx, &parsedTx)
		if err != nil {
			logrus.WithError(err).Errorln("Unmarshal AssignmentBFTTx failed")
			return false, err
		}
	}
	return true, nil
}

func authenticateBftTx(tx []byte) (DefaultBFTTxWrapper, error) {
	var parsedTx DefaultBFTTxWrapper
	err := json.Unmarshal(tx, &parsedTx)
	if err != nil {
		return DefaultBFTTxWrapper{}, err
	}
	// TODO: Verify that the sender valid is whitelist
	// Check by public key in whitelist epoch
	return parsedTx, nil
}
