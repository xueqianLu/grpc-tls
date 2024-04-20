/*
This package defines the messages for client messages
*/

package message

import (
	pb "chainhealth/src/proto/proto/communication"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/vmihailenco/msgpack"
	"net/http"
)

/*
Type of read requests
*/
type TypeOfRead int

const (
	Read_SeqByCid TypeOfRead = 1 //read all the seq numbers and transactions of one client
	Reqd_TDH2Key  TypeOfRead = 2 //read user own status (for user)
	Read_Cipher   TypeOfRead = 3 //read cipher
	Read_Plain    TypeOfRead = 4
)

var TypeOfRead_name = map[string]TypeOfRead{
	"1": Read_SeqByCid,
	"2": Reqd_TDH2Key,
	"3": Read_Cipher,
	"4": Read_Plain,
}

/*
Type of write requests
*/
type TypeOfWrite int

const (
	Write_Plain     TypeOfWrite = 0 //plain type where the data is simply written on the blockchain
	Write_Cipher    TypeOfWrite = 1
	Write_AddACL    TypeOfWrite = 2
	Write_DeleteACL TypeOfWrite = 3
)

var TypeOfWrite_name = map[int]TypeOfWrite{
	0: Write_Plain,
	1: Write_Cipher,
	2: Write_AddACL,
	3: Write_DeleteACL,
}

/*
Type of dynamic node requests
*/
type TypeOfDynamicNode int

const (
	Dynamic_Join  TypeOfDynamicNode = 0 //plain type where the data is simply written on the blockchain
	Dynamic_Leave TypeOfDynamicNode = 1
)

var TypeOfDynamicNode_name = map[int]TypeOfDynamicNode{
	0: Dynamic_Join,
	1: Dynamic_Leave,
}

/*
Type of admin request
*/
type TypeOfAdminRequest int

const (
	Register TypeOfAdminRequest = 0
	Freeze   TypeOfAdminRequest = 1
	Unfreeze TypeOfAdminRequest = 2
	Logout   TypeOfAdminRequest = 3
)

var TypeOfAdminRequest_name = map[int]TypeOfAdminRequest{
	0: Register,
	1: Freeze,
	2: Unfreeze,
	3: Logout,
}

type ClientRequest struct {
	Type pb.MessageType
	ID   int64  `json:"ccid"`
	OP   []byte // Message c.
	PK   []byte // Public key.(not using)
	TS   int64  // Timestamp
}

/*
ClientReply.Msg
*/
type MsgInfo struct {
	Seq  int64  `json:"seq"`
	TxID string `json:"txid"`
}

/*
Reply messages
*/
type ClientReply struct {
	Type   pb.MessageType `json:"type"`
	Source int64          `json:"source"`
	Result bool           `json:"result"`
	Re     string         `json:"reply"`
	Msg    []byte         `json:"msg"`
	Share  []byte         `json:"share"`
}

/*
Get String to ClientRequest
*/
func (r *ClientRequest) String() string {
	var p string
	OP := DeserializeWriteRequest(r.OP)
	p = fmt.Sprintf("ID: %d, Operation: %s, Timestamp: %d", r.ID, OP.String(), r.TS)
	return p
}

func (r *ClientRequest) StringExceptUID() string {
	var p string
	OP := DeserializeWriteRequest(r.OP)
	p = fmt.Sprintf("ID: %d, Operation: %s, Timestamp: %d", r.ID, OP.StringExceptUID(), r.TS)
	return p
}

/*
Serialize ClientRequest
*/
func (r *ClientRequest) Serialize() ([]byte, error) {
	jsons, err := msgpack.Marshal(r)
	if err != nil {
		return []byte(""), err
	}
	return jsons, nil
}

/*
Deserialize []byte to ClientRequest
*/
func DeserializeClientRequest(input []byte) ClientRequest {
	var clientRequest = new(ClientRequest)
	msgpack.Unmarshal(input, &clientRequest)
	return *clientRequest
}

/*
Serialize ClientReply
*/
func (r *ClientReply) Serialize() ([]byte, error) {
	jsons, err := msgpack.Marshal(r)
	if err != nil {
		return []byte(""), err
	}
	return jsons, nil
}

/*
Deserialize []byte to ClientReply
*/
func DeserializeClientReply(input []byte) ClientReply {
	var clientReply = new(ClientReply)
	msgpack.Unmarshal(input, &clientReply)
	return *clientReply
}

/*
Serialize MsgInfo
*/
func (r *MsgInfo) Serialize() ([]byte, error) {
	jsons, err := msgpack.Marshal(r)
	if err != nil {
		return []byte(""), err
	}
	return jsons, nil
}

/*
Deserialize []byte to ClientRequest
*/
func DeserializeMsgInfo(input []byte) MsgInfo {
	var msgInfo = new(MsgInfo)
	msgpack.Unmarshal(input, &msgInfo)
	return *msgInfo
}

/*
Read request
*/
type ReadRequest struct {
	RType      string `json:"rType"`      // Type of the read request
	DataOwner  int64  `json:"DataOwner"`  // Owner of the data. Used for read type 2
	BeginTS    int64  `json:"BeginTS"`    // beginning timestamp
	EndTS      int64  `json:"EndTS"`      // ending timestamp
	RequestTag string `json:"RequestTag"` //tag that replicas reply for a tx
	TS         int64  `json:"CurTS"`      // current timestamp
	TxID       string `json:"txID"`       //信息id，即唯一标识符
	PK         string `json:"pk"`
}

/*
Read Reply
*/
type ReadReply struct {
	SEQ  []int           `json:"Sequence"` //seq number which the request belongs
	TXID []string        `json:"TXID"`
	OPS  []ClientRequest `json:"Requests"` // Client requests
}

/*
Used for server APIs. Statistics of the system
*/
type Statistics struct {
	View   int `json:"View"`
	Leader int `json:"Leader"`
	Conf   int `json:"Conf"`
	CurSeq int `json:"Seq"`
}

/*
Serialize ReadReply to bytes. Used for json read types
*/
func (r *ReadReply) Serialize() ([]byte, error) {
	jsons, err := msgpack.Marshal(r)
	if err != nil {
		return []byte(""), err
	}
	return jsons, nil
}

/*
Deserialize bytes to ReadReply. Used for json read types
*/
func DeserializeReadReply(input []byte) ReadReply {
	var readReply = new(ReadReply)
	msgpack.Unmarshal(input, &readReply)
	return *readReply
}

/*
Serialize statistics to bytes. Used for server APIs.
*/
func (r *Statistics) Serialize() ([]byte, error) {
	jsons, err := msgpack.Marshal(r)
	if err != nil {
		return []byte(""), err
	}
	return jsons, nil
}

/*
Deserialize bytes to statistics. Used for server APIs.
*/
func DeserializeStatistics(input []byte) Statistics {
	var statistics = new(Statistics)
	msgpack.Unmarshal(input, &statistics)
	return *statistics
}

/*
Serialize ReadRequest to bytes. Used for client APIs.
*/
func (r *ReadRequest) Serialize() ([]byte, error) {
	jsons, err := msgpack.Marshal(r)
	if err != nil {
		return []byte(""), err
	}
	return jsons, nil
}

/*
Deserialize bytes to ReadRequest. Used for client APIs.
*/
func DeserializeReadRequest(input []byte) ReadRequest {
	var readRequest = new(ReadRequest)
	msgpack.Unmarshal(input, &readRequest)
	return *readRequest
}

/*
Printable function for client requests
*/
func GetOPSString(r pb.RawMessage) string {
	rawMessage := DeserializeMessageWithSignature(r.GetMsg())
	CR := DeserializeClientRequest(rawMessage.Msg)
	p := fmt.Sprintf("Regular Transaction: %s: Sig: %s", CR.String(), rawMessage.Sig)
	return p
}

/*
Printable function for regular transactions
*/
func (r *ReplicaMessage) String() string {
	p := fmt.Sprintf("Seq: %d, View: %d, Leader: %d", r.Seq, r.View, r.Source)

	for i := 0; i < len(r.OPS); i++ {
		p += "\t" + GetOPSString(r.OPS[i]) + "\n"
	}

	return p
}

/*
Printable function for membership transaction. Used for dynanamic membership and static membership.
*/
func (r *JoinMessage) String() string {
	p := fmt.Sprintf("MembershipType: %d", r.MemType)
	return p
}

/*
Write request. Currently used for the supply chain demo.
*/
type WriteRequest struct {
	WType int    `json:"wType"` // Type of the write request
	UID   string `json:"UID"`   // ID of the product
	OP    []byte `json:"OP"`    // contents of the data
}

/*
Serialize WriteRequest to bytes. Used for supply chain app.
*/
func (r *WriteRequest) Serialize() ([]byte, error) {
	jsons, err := msgpack.Marshal(r)
	if err != nil {
		return []byte(""), err
	}
	return jsons, nil
}

/*
Deserialize bytes to WriteRequest. Used for supply chain app.
*/
func DeserializeWriteRequest(input []byte) WriteRequest {
	var writeRequest = new(WriteRequest)
	msgpack.Unmarshal(input, &writeRequest)
	return *writeRequest
}

func (r *WriteRequest) String() string {
	p := fmt.Sprintf("{wType:%d, UID:%s, OP:%s}", r.WType, r.UID, r.OP)
	return p
}

func (r *WriteRequest) StringExceptUID() string {
	p := fmt.Sprintf("{wType:%d,OP:%s}", r.WType, r.OP)
	return p
}

/*
Write request. Currently used for the supply chain demo.
*/
//type WriteInfo struct {
//	UserID string   `json:"userID"`
//	Txid string   `json:"Txid"`
//	Info   []byte   `json:"info"`
//	Hash   []byte   `json:"hash"`
//	Acl    [][]byte `json:"acl"`
//}

/*
Serialize WriteInfo to bytes. Used for supply chain app.
*/
//func (r *WriteInfo) Serialize() ([]byte, error) {
//	msg, err := msgpack.Marshal(r)
//	if err != nil {
//		return []byte(""), err
//	}
//	return msg, nil
//}

/*
Deserialize bytes to WriteRequest. Used for supply chain app.
*/
//func DeserializeWriteInfo(input []byte) WriteInfo {
//	var writeInfo = new(WriteInfo)
//	msgpack.Unmarshal(input, &writeInfo)
//	return *writeInfo
//}

//type ReadInfo struct {
//	UID string
//	C   string //c
//	PK  string
//}
//
//func (r *ReadInfo) Serialize() ([]byte, error) {
//	msg, err := msgpack.Marshal(r)
//	if err != nil {
//		return []byte(""), err
//	}
//	return msg, nil
//}
//
//func DeserializeReadInfo(input []byte) ReadInfo {
//	var readInfo = new(ReadInfo)
//	msgpack.Unmarshal(input, &readInfo)
//	return *readInfo
//}

type CipherResponse struct {
	C      []byte         `json:"c"`
	Hash   []byte         `json:"hash"`
	Shares map[int][]byte `json:"shares"`
}

func (r *CipherResponse) Serialize() ([]byte, error) {
	msg, err := msgpack.Marshal(r)
	if err != nil {
		return []byte(""), err
	}
	return msg, nil
}

func DeserializeCipherResponse(input []byte) CipherResponse {
	var msg = new(CipherResponse)
	msgpack.Unmarshal(input, &msg)
	return *msg
}

type AdminRequest struct {
	Type   int    `json:"type"`
	UserID int64  `json:"userID"`
	PK     string `json:"pk"`
}

func (r *AdminRequest) Serialize() ([]byte, error) {
	msg, err := msgpack.Marshal(r)
	if err != nil {
		return []byte(""), err
	}
	return msg, nil
}

func DeserializeAdminRequest(input []byte) AdminRequest {
	var msg = new(AdminRequest)
	msgpack.Unmarshal(input, &msg)
	return *msg
}

// API message struct
type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func ReturnMsg(ctx *gin.Context, code int, data interface{}) {
	ctx.JSON(code, Response{
		Code: code,
		Msg:  http.StatusText(code),
		Data: data,
	})

	return
}

type WriteInfo struct {
	UserID int64    `json:"userID"`
	Txid   string   `json:"txID"`
	Info   string   `json:"info"`
	Hash   string   `json:"hash"`
	Acl    []string `json:"acl"`
}

func (r *WriteInfo) Serialize() ([]byte, error) {
	msg, err := msgpack.Marshal(r)
	if err != nil {
		return []byte(""), err
	}
	return msg, nil
}

func DeserializeWriteInfo(input []byte) WriteInfo {
	var request = new(WriteInfo)
	msgpack.Unmarshal(input, &request)
	return *request
}

type ServerAPIRequest struct {
	Nid string `json:"nid"`
	Uid string `json:"uid"`
}

type BlockIndex struct {
	SSeq int64 `json:"sseq"`
	EEeq int64 `json:"eseq"`
}
