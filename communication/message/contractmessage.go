package message

import (
	"github.com/vmihailenco/msgpack"
)

type TypeOfContract int

const (
	NoContract TypeOfContract = 0 //no contract
	EVM        TypeOfContract = 1 // EVM contract
)

type TypeOfContractRequest int

const (
	DeployContract   TypeOfContractRequest = 1 //deploy contract
	CallContract     TypeOfContractRequest = 2 //call contract
	FreezeContract   TypeOfContractRequest = 3 //freeze contract
	UnfreezeContract TypeOfContractRequest = 4 //unfreeze contract
	WithdrawContract TypeOfContractRequest = 5 //deprecate contract
)

var TypeOfContractRequest_name = map[int]TypeOfContractRequest{
	1: DeployContract,
	2: CallContract,
	3: FreezeContract,
	4: UnfreezeContract,
	5: WithdrawContract,
}

type TypeOfContractState int

const (
	Contract_Normal   TypeOfContractState = 0
	Contract_Freezing TypeOfContractState = 1
	Contract_Withdrew TypeOfContractState = 2
)

var TypeOfContractState_name = map[int]TypeOfContractState{
	0: Contract_Normal,
	1: Contract_Freezing,
	2: Contract_Withdrew,
}

type ContractContent struct {
	TxID       string
	Method     string
	Params     []string
	Version    string
	CodeString string
	ABI        string
	State      TypeOfContractState
}

type ContractRequest struct {
	UserID    string
	CType     int
	CContents ContractContent
}

type ContractReply struct {
	ContractTxID string
	Method       string
	Params       []string
	Content      interface{}
	Err          error
}

/*
Deserialize bytes to ContractReply.
*/
func DeserializeContractReply(input []byte) ContractReply {
	var contractReply = new(ContractReply)
	msgpack.Unmarshal(input, &contractReply)
	return *contractReply
}

/*
Serialize ContractContent to bytes.
*/
func (r *ContractReply) Serialize() ([]byte, error) {
	jsons, err := msgpack.Marshal(r)
	if err != nil {
		return []byte(""), err
	}
	return jsons, nil
}

/*
Serialize ContractContent to bytes.
*/
func (r *ContractContent) Serialize() ([]byte, error) {
	jsons, err := msgpack.Marshal(r)
	if err != nil {
		return []byte(""), err
	}
	return jsons, nil
}

/*
Deserialize bytes to ContractContent.
*/
func DeserializeContractContent(input []byte) ContractContent {
	var ContractContent = new(ContractContent)
	msgpack.Unmarshal(input, &ContractContent)
	return *ContractContent
}

/*
Serialize ContractContent to bytes.
*/
func (r *ContractRequest) Serialize() ([]byte, error) {
	jsons, err := msgpack.Marshal(r)
	if err != nil {
		return []byte(""), err
	}
	return jsons, nil
}

/*
Deserialize bytes to ContractContent.
*/
func DeserializeContractRequest(input []byte) ContractRequest {
	var contractRequest = new(ContractRequest)
	msgpack.Unmarshal(input, &contractRequest)
	return *contractRequest
}
