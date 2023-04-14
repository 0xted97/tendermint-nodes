package services

import (
	"crypto/elliptic"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/url"
	"os"
	"regexp"
	"strconv"

	"github.com/dgraph-io/badger"
	"github.com/me/dkg-node/config"
	abcitypes "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/version"
)

type VerifyPair struct {
	Verifier   string `json:"verifier"`
	VerifierID string `json:"verifierID"`
}

type Point struct {
	X big.Int
	Y big.Int
}

type KeyAssignmentPublic struct {
	Index     int
	PublicKey []byte
	Threshold int
	Verifiers map[string]string // Verifier => VerifierID
}

type State struct {
	LastUnassignedIndex uint `json:"last_unassigned_index"`
	LastCreatedIndex    uint
	NewKeyAssignments   []KeyAssignmentPublic `json:"new_key_assignments"`
}

type ABCIApp struct {
	db    *badger.DB
	state *State
}

func randomString(n int) string {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return base64.URLEncoding.EncodeToString(b)
}

func (app *ABCIApp) getKeyIndex(verifierID string, verifier string) int {
	for i, ka := range app.state.NewKeyAssignments {
		if ka.Verifiers[verifier] == verifierID {
			return i
		}
	}
	return -1
}

func generateRandomKeyAssignments(n int, threshold int) ([]KeyAssignmentPublic, error) {
	assignments := make([]KeyAssignmentPublic, n)
	curve := elliptic.P256()

	for i := 0; i < n; i++ {
		_, x, y, err := elliptic.GenerateKey(curve, rand.Reader)
		if err != nil {
			return nil, fmt.Errorf("failed to generate key pair: %v", err)
		}
		publicKey := elliptic.Marshal(curve, x, y)

		verifiers := make(map[string]string)
		// verifiers[randomString(10)] = randomString(10)
		assignments[i] = KeyAssignmentPublic{
			Index:     i,
			PublicKey: publicKey,
			Threshold: threshold,
			Verifiers: verifiers,
		}
	}

	return assignments, nil
}

func (ABCIService) NewABCIApp() *ABCIApp {
	assignments, err := generateRandomKeyAssignments(10, 3) // Adjust threshold as needed
	if err != nil {
		fmt.Printf("Failed to generate random key assignments: %v\n", err)
		return nil
	}

	db, err := badger.Open(badger.DefaultOptions(config.GlobalConfig.DatabasePath))
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open badger db: %v", err)
		os.Exit(1)
	}
	defer db.Close()

	return &ABCIApp{
		db: db,
		state: &State{
			LastUnassignedIndex: 0,
			LastCreatedIndex:    10,
			NewKeyAssignments:   assignments,
		},
	}
}

func (app *ABCIApp) Info(req abcitypes.RequestInfo) abcitypes.ResponseInfo {
	return abcitypes.ResponseInfo{
		Version: version.ABCIVersion,
	}
}

func (ABCIApp) SetOption(req abcitypes.RequestSetOption) abcitypes.ResponseSetOption {
	return abcitypes.ResponseSetOption{}
}

func (app *ABCIApp) DeliverTx(req abcitypes.RequestDeliverTx) abcitypes.ResponseDeliverTx {
	code := app.isValid(req.Tx)
	if code != 0 {
		return abcitypes.ResponseDeliverTx{Code: code}
	}

	// Full key
	if app.state.LastUnassignedIndex >= app.state.LastCreatedIndex {
		return abcitypes.ResponseDeliverTx{Code: 1, Log: "No more key assignments available"}
	}
	// Assign key
	queryStr := string(req.Tx)
	queryParams, err := url.ParseQuery(queryStr)
	if err != nil {
		return abcitypes.ResponseDeliverTx{Code: 1}
	}
	verifierID := queryParams.Get("verifierID")
	verifier := queryParams.Get("verifier")
	fmt.Printf("verifier: %v\n", verifier)
	fmt.Printf("verifierID: %v\n", verifierID)
	index := app.getKeyIndex(verifierID, verifier)
	if index >= 0 {
		return abcitypes.ResponseDeliverTx{Code: 1, Log: "Key already assigned to verifier"}
	}
	// Assign key to verifierID and verifier
	app.state.NewKeyAssignments[app.state.LastUnassignedIndex].Verifiers[verifier] = verifierID
	app.state.LastUnassignedIndex = app.state.LastUnassignedIndex + 1

	return abcitypes.ResponseDeliverTx{Code: abcitypes.CodeTypeOK}
}

func (app *ABCIApp) CheckTx(req abcitypes.RequestCheckTx) abcitypes.ResponseCheckTx {
	code := app.isValid(req.Tx)
	if code != 0 {
		return abcitypes.ResponseCheckTx{
			Code: code,
			Log:  "transaction is invalid",
		}
	}
	// If the transaction is valid, return a success response
	return abcitypes.ResponseCheckTx{
		Code: code,
		Log:  "transaction is valid",
	}
}

func (app *ABCIApp) Commit() abcitypes.ResponseCommit {
	return abcitypes.ResponseCommit{Data: []byte{}}
}

func (app *ABCIApp) Query(reqQuery abcitypes.RequestQuery) (resQuery abcitypes.ResponseQuery) {
	switch reqQuery.Path {
	case "/KeyAssignment":
		index := string(reqQuery.Data)
		if indexInt, err := strconv.Atoi(index); err == nil && indexInt >= 0 && indexInt < len(app.state.NewKeyAssignments) {
			resQuery.Key = reqQuery.Data
			keyAssignmentJSON, err := json.Marshal(app.state.NewKeyAssignments[indexInt])
			fmt.Printf("app.state.NewKeyAssignments[indexInt]: %v\n", app.state.NewKeyAssignments[indexInt])
			if err != nil {
				resQuery.Code = 1
				resQuery.Log = fmt.Sprintf("error marshalling key assignment: %v", err)
			} else {
				resQuery.Value = keyAssignmentJSON
				resQuery.Code = 0
				resQuery.Log = "success"
			}
		} else {
			resQuery.Code = 1
			resQuery.Log = "invalid index"
		}
		break
	case "/GetIndexesFromVerifierID":
		queryStr := string(reqQuery.Data)
		queryParams, err := url.ParseQuery(queryStr)
		verifier := queryParams.Get("verifier")
		verifierID := queryParams.Get("verifierID")
		if err != nil {
			resQuery.Code = 1
			resQuery.Log = fmt.Sprintf("error unmarshalling query: %v", err)
		}
		keyIndex := app.getKeyIndex(verifierID, verifier)
		fmt.Printf("keyIndex: %v\n", keyIndex)
		if keyIndex < 0 {
			resQuery.Code = 1
			resQuery.Log = fmt.Sprintf("error key not found")
		} else {
			keyAssignmentJSON, err := json.Marshal(app.state.NewKeyAssignments[keyIndex])
			if err != nil {
				resQuery.Code = 1
				resQuery.Log = fmt.Sprintf("error marshalling key assignment: %v", err)
			} else {
				resQuery.Value = keyAssignmentJSON
				resQuery.Code = 0
				resQuery.Log = "success"
			}

		}
		break
	default:
		resQuery.Code = 1
		resQuery.Log = "unknown query path"
	}
	return
}

func (ABCIApp) InitChain(req abcitypes.RequestInitChain) abcitypes.ResponseInitChain {
	return abcitypes.ResponseInitChain{}
}

func (app *ABCIApp) BeginBlock(req abcitypes.RequestBeginBlock) abcitypes.ResponseBeginBlock {
	println("BeginBlock: ", req.Header.Height)
	return abcitypes.ResponseBeginBlock{}
}

func (ABCIApp) EndBlock(req abcitypes.RequestEndBlock) abcitypes.ResponseEndBlock {
	return abcitypes.ResponseEndBlock{}
}

func (ABCIApp) ListSnapshots(req abcitypes.RequestListSnapshots) abcitypes.ResponseListSnapshots {
	return abcitypes.ResponseListSnapshots{}
}

func (ABCIApp) OfferSnapshot(req abcitypes.RequestOfferSnapshot) abcitypes.ResponseOfferSnapshot {
	return abcitypes.ResponseOfferSnapshot{}
}

func (ABCIApp) LoadSnapshotChunk(abcitypes.RequestLoadSnapshotChunk) abcitypes.ResponseLoadSnapshotChunk {
	return abcitypes.ResponseLoadSnapshotChunk{}
}

func (ABCIApp) ApplySnapshotChunk(abcitypes.RequestApplySnapshotChunk) abcitypes.ResponseApplySnapshotChunk {
	return abcitypes.ResponseApplySnapshotChunk{}
}

func (app *ABCIApp) isValid(tx []byte) (code uint32) {
	queryStr := string(tx)
	queryParams, err := url.ParseQuery(queryStr)
	if err != nil {
		panic(err)
	}
	verifierID := queryParams.Get("verifierID")
	// verifier := queryParams.Get("verifier")
	// Validate the verifierID is an email
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(verifierID) {
		code = 1
	}
	return code
}
