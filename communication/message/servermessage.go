/*
This package defines all the messages types, serialization and deserialization of the messages.
*/

package message

import (
	"chainhealth/src/cryptolib"
	pb "chainhealth/src/proto/proto/communication"
	"github.com/vmihailenco/msgpack"
)

/*
Type of message. Used for ByteSend in sender.go
*/
type TypeOfMessage int

const (
	ReplicaMsg    TypeOfMessage = 0
	CheckpointMsg TypeOfMessage = 1
	VCMsg         TypeOfMessage = 2
	JoinMsg       TypeOfMessage = 3
)

/*
Type of replica joining the system.
*/
type TypeOfMembership int

const (
	ExiRepCleanDB TypeOfMembership = 0 // Existing replica re-joins with a clean DB.
	ExiRepDB      TypeOfMembership = 1 // Existing replica re-joins with data from past views.
	StateTransfer TypeOfMembership = 2 // Sync with one node
	Join          TypeOfMembership = 3 //
	Leave         TypeOfMembership = 4 //
)

var TypeOfMembership_name = map[int]TypeOfMembership{
	0: ExiRepCleanDB,
	1: ExiRepDB,
	2: StateTransfer,
	3: Join,
	4: Leave,
}

/*
Type of user status
*/
type TypeOfUserStatus int

const (
	NotExist TypeOfUserStatus = -1
	Formal   TypeOfUserStatus = 0
	Freezing TypeOfUserStatus = 1
	NewGlory TypeOfUserStatus = 2
)

var TypeOfUserStatus_int = map[TypeOfUserStatus]int{
	NotExist: -1,
	Formal:   0,
	Freezing: 1,
	NewGlory: 2,
}

/*
Type of user role
*/
type TypeOfUserRole int

const (
	ManageUser     TypeOfUserRole = 0
	ManageNode     TypeOfUserRole = 1
	ManageTx       TypeOfUserRole = 2
	ManageContract TypeOfUserRole = 3
)

var TypeOfUserRole_int = map[TypeOfUserRole]int{
	ManageUser:     0,
	ManageNode:     1,
	ManageTx:       2,
	ManageContract: 3,
}

type TypeOfServerAPI int

const (
	GetChainInfo     TypeOfServerAPI = 0
	GetNodeInfo      TypeOfServerAPI = 1
	GetBlockInfo     TypeOfServerAPI = 2
	GetTDH2PublicKey TypeOfServerAPI = 3
)

/*
PREPARE certificates and COMMIT certificates
*/
type Cer struct {
	Msgs [][]byte
}

/*
reply seq number and request hash to client
*/
type RequestTag struct {
	Seq  int
	Hash string
}

/*
Consensus messages: PRE-PREPARE, PREPARE, and COMMIT messages
RegisterInfo is used for client registration
*/
type ReplicaMessage struct {
	Mtype  pb.MessageType
	Seq    int
	Source int64
	View   int
	OPS    []pb.RawMessage
	Hash   []byte
	Hashes map[string]int
	Time   int64
}

/*
VIEWCHANGE and NEWVIEW messages
*/
type ViewChangeMessage struct {
	Mtype  pb.MessageType
	View   int
	Seq    int
	Source int64
	C      []MessageWithSignature       //CheckpointMessage
	P      map[int]Cer                  //Prepare certificate
	O      map[int]MessageWithSignature // Operations. Used in new-view message only
	V      []MessageWithSignature       // A quorum of view-change messages. Used in new-view message only
}

// Join and catchup messages
type JoinMessage struct {
	Mtype         pb.MessageType
	MemType       TypeOfMembership
	View          int
	Source        int64
	Seq           int
	Hash          []byte
	PCer          map[int]Cer
	ClientMap     map[int64]int64
	LastStableSeq int
	//ClientSeqStore map[int64][]int        // store the client id and the corresponding seq number, for read only
	//SeqClientStore map[int][]int64        // store the seq number and the corresponding client id, for read only
	//SeqTimeStore   map[int][]int64        // store the seq number and its min timestamp and max timestamp, for read only
	C        []MessageWithSignature //CheckpointMessage
	Store    map[string]ReplicaMessage
	SeqStore map[int]string
}

type CheckpointMessage struct {
	Seq    int
	Source int64
	Hash   []byte
}

// Message with signautre
type MessageWithSignature struct {
	Msg     []byte `json:"msg"`
	Version int    `json:"version"`
	Sig     []byte `json:"sig"`
}

type StateInfo struct {
	Num        int
	LastStable int
	C          []MessageWithSignature
}

type LargeState struct {
	State map[int][]byte
}

/*
Serialize StateInfo
*/
func (r *StateInfo) Serialize() ([]byte, error) {
	jsons, err := msgpack.Marshal(r)
	if err != nil {
		return []byte(""), err
	}
	return jsons, nil
}

/*
Deserialize []byte to StateInfo
*/
func DeserializeStateInfo(input []byte) StateInfo {
	var stateInfo = new(StateInfo)
	msgpack.Unmarshal(input, &stateInfo)
	return *stateInfo
}

/*
Serialize LargeState
*/
func (r *LargeState) Serialize() ([]byte, error) {
	jsons, err := msgpack.Marshal(r)
	if err != nil {
		return []byte(""), err
	}
	return jsons, nil
}

/*
Deserialize []byte to LargeState
*/
func DeserializeLargeState(input []byte) LargeState {
	var largeState = new(LargeState)
	msgpack.Unmarshal(input, &largeState)
	return *largeState
}

/*
Serialize MessageWithSignature
*/
func (r *MessageWithSignature) Serialize() ([]byte, error) {
	jsons, err := msgpack.Marshal(r)
	if err != nil {
		return []byte(""), err
	}
	return jsons, nil
}

/*
Deserialize []byte to MessageWithSignature
*/
func DeserializeMessageWithSignature(input []byte) MessageWithSignature {
	var messageWithSignature = new(MessageWithSignature)
	msgpack.Unmarshal(input, &messageWithSignature)
	return *messageWithSignature
}

/*
Serialize CheckpointMessage
*/
func (r *CheckpointMessage) Serialize() ([]byte, error) {
	jsons, err := msgpack.Marshal(r)
	if err != nil {
		return []byte(""), err
	}
	return jsons, nil
}

/*
Deserialize []byte to CheckpointMessage
*/
func DeserializeCheckpointMessage(input []byte) CheckpointMessage {
	var checkpointMessage = new(CheckpointMessage)
	msgpack.Unmarshal(input, &checkpointMessage)
	return *checkpointMessage
}

/*
Serialize JoinMessage
*/
func (r *JoinMessage) Serialize() ([]byte, error) {
	jsons, err := msgpack.Marshal(r)
	if err != nil {
		return []byte(""), err
	}
	return jsons, nil
}

/*
Deserialize []byte to JoinMessage
*/
func DeserializeJoinMessage(input []byte) JoinMessage {
	var joinMessage = new(JoinMessage)
	msgpack.Unmarshal(input, &joinMessage)
	return *joinMessage
}

func (r *RequestTag) Serialize() ([]byte, error) {
	jsons, err := msgpack.Marshal(r)
	if err != nil {
		return []byte(""), err
	}
	return jsons, nil
}

func DeserializeRequestTag(input []byte) RequestTag {
	var replyTag = new(RequestTag)
	msgpack.Unmarshal(input, &replyTag)
	return *replyTag
}

/*
Serialize ReplicaMessage
*/
func (r *ReplicaMessage) Serialize() ([]byte, error) {
	jsons, err := msgpack.Marshal(r)
	if err != nil {
		return []byte(""), err
	}
	return jsons, nil
}

/*
Convert an Prepare message to Commit message by changing the source node id only
Input

	id: id of the node (int64 type)
*/
func (r *ReplicaMessage) EchoToReady(id int64) {
	r.Mtype = pb.MessageType_COMMIT
	r.Source = id
}

/*
Get hash of the entire batch
*/
func (r *ReplicaMessage) GetMsgHash() []byte {
	if len(r.OPS) == 0 {
		return []byte("")
	}
	return cryptolib.GenBatchHash(r.OPS)
}

/*
Get number of requests in a batch
*/
func (r *ReplicaMessage) OPSLen() int {
	return len(r.OPS)
}

/*
Get requests in the batch
*/
func (r *ReplicaMessage) GetOPS() []pb.RawMessage {
	return r.OPS
}

/*
Get hashes (a list of hash) of the requests in the batch
*/
func (r *ReplicaMessage) GetOPSHash() [][]byte {
	l := len(r.OPS)
	result := make([][]byte, l, l)
	for i := 0; i < l; i++ {
		result[i] = cryptolib.GenHash(r.OPS[i].GetMsg())
	}
	return result
}

/*
Deserialize []byte to ReplicaMessage
*/
func DeserializeReplicaMessage(input []byte) ReplicaMessage {
	var replicaMessage = new(ReplicaMessage)
	msgpack.Unmarshal(input, &replicaMessage)
	return *replicaMessage
}

/*
Deserialize []byte to ViewChangeMessage
*/
func DeserializeViewChangeMessage(input []byte) ViewChangeMessage {
	var viewChangeMessage = new(ViewChangeMessage)
	msgpack.Unmarshal(input, &viewChangeMessage)
	return *viewChangeMessage
}

/*
Serialize ViewChangeMessage
*/
func (r *ViewChangeMessage) Serialize() ([]byte, error) {
	jsons, err := msgpack.Marshal(r)
	if err != nil {
		return []byte(""), err
	}
	return jsons, nil
}

/*
Add a message to the certificate
*/
func (r *Cer) Add(msg []byte) {
	r.Msgs = append(r.Msgs, msg)
}

/*
Get messages in a certificate
*/
func (r *Cer) GetMsgs() [][]byte {
	return r.Msgs
}

/*
Get the number of messages in the certiicate
*/
func (r *Cer) Len() int {
	return len(r.Msgs)
}

/*
Create MessageWithSignature where the signature of each message (in bytes) is attached.
Input

	id: node id (int64 type)
	msg: input message, serialized []byte format

Output

	[]byte: serialized MessageWithSignature
	error: nil if no error
*/
func SerializeWithSignature(id int64, msg []byte) ([]byte, error) {
	request := MessageWithSignature{
		Msg: msg,
		Sig: cryptolib.GenSig(id, msg),
	}

	requestSer, err := request.Serialize()
	if err != nil {
		return []byte(""), err
	}
	return requestSer, err
}

/*
Create MessageWithSignature where the signature of each message (in ReplicaMessage) is attached.
The ReplicaMessage is first serialized into bytes and then signed.
Input

	tmpmsg: ReplicaMessage

Output

	MessageWithSignature: the struct
*/
func CreateMessageWithSig(tmpmsg ReplicaMessage) MessageWithSignature {
	tmpmsgSer, err := tmpmsg.Serialize()
	if err != nil {
		var emptymsg MessageWithSignature
		return emptymsg
	}

	op := MessageWithSignature{
		Msg: tmpmsgSer,
		Sig: cryptolib.GenSig(tmpmsg.Source, tmpmsgSer),
	}
	return op
}

/*
Block Information
*/
type TxInfo struct {
	Mtype  pb.MessageType
	Txid   string
	RawMsg pb.RawMessage
}

type BlockInfo struct {
	Seq    int      `json:"height"`  // height of the block
	Hash   []byte   `json:"hash"`    // hash of the block
	Len    int      `json:"numOfTx"` // number of tx
	Time   int64    `json:"time"`    // time
	Txs    []TxInfo `json:"txs"`
	Leader string   `json:"leader"`
	View   string   `json:"view"`
}

type Blocks struct {
	BlockList []BlockInfo `json:"blocks"`
}

/*
Serialize BlockInfo to bytes.
*/
func (r *BlockInfo) Serialize() ([]byte, error) {
	jsons, err := msgpack.Marshal(r)
	if err != nil {
		return []byte(""), err
	}
	return jsons, nil
}

/*
Deserialize bytes to BlockInfo.
*/
func DeserializeBlockInfo(input []byte) BlockInfo {
	var blockInfo = new(BlockInfo)
	msgpack.Unmarshal(input, &blockInfo)
	return *blockInfo
}

/*
Serialize Blocks to bytes.
*/
func (r *Blocks) Serialize() ([]byte, error) {
	jsons, err := msgpack.Marshal(r)
	if err != nil {
		return []byte(""), err
	}
	return jsons, nil
}

/*
Deserialize bytes to Blocks.
*/
func DeserializeBlocks(input []byte) Blocks {
	var blocks = new(Blocks)
	msgpack.Unmarshal(input, &blocks)
	return *blocks
}

/*
Chain Information
*/
type ChainInfo struct {
	NodeNumber    int   `json:"nodeNumber"`
	Height        int   `json:"height"`
	TxNumber      int64 `json:"numOfTx"`
	AccountNumber int64 `json:"numOfAccount"`
}

/*
Deserialize bytes to BlockInfo.
*/
func DeserializeChainInfo(input []byte) ChainInfo {
	var chainInfo = new(ChainInfo)
	msgpack.Unmarshal(input, &chainInfo)
	return *chainInfo
}

/*
Serialize Blocks to bytes.
*/
func (r *ChainInfo) Serialize() ([]byte, error) {
	jsons, err := msgpack.Marshal(r)
	if err != nil {
		return []byte(""), err
	}
	return jsons, nil
}

/*
Chain Information
*/
type NodeInfos struct {
	Nodes []NodeInfo
}

type NodeInfo struct {
	ID string `json:"id"`
	IP string `json:"ip"`
}

/*
Deserialize bytes to BlockInfo.
*/
func DeserializeNodeInfos(input []byte) NodeInfos {
	var nodeInfos = new(NodeInfos)
	msgpack.Unmarshal(input, &nodeInfos)
	return *nodeInfos
}

/*
Serialize Blocks to bytes.
*/
func (r *NodeInfos) Serialize() ([]byte, error) {
	jsons, err := msgpack.Marshal(r)
	if err != nil {
		return []byte(""), err
	}
	return jsons, nil
}

/*
Dynamic node information
*/
type DynamicNodeInfo struct {
	DType  TypeOfDynamicNode `json:"dtype"`
	NodeID string            `json:"nodeID"`
	IP     string            `json:"ip"`
	Port   string            `json:"port"`
}

/*
Deserialize bytes to DynamicNodeInfo.
*/
func DeserializeDynamicNodeInfo(input []byte) DynamicNodeInfo {
	var dynamicNodeInfo = new(DynamicNodeInfo)
	msgpack.Unmarshal(input, &dynamicNodeInfo)
	return *dynamicNodeInfo
}

/*
Serialize DynamicNodeInfo to bytes.
*/
func (r *DynamicNodeInfo) Serialize() ([]byte, error) {
	jsons, err := msgpack.Marshal(r)
	if err != nil {
		return []byte(""), err
	}
	return jsons, nil
}
