/*
Sender functions.
It implements all sending functions for clients
*/

package sender

import (
	"chainhealth/src/communication"
	"chainhealth/src/communication/message"
	"chainhealth/src/consts"
	"chainhealth/src/cryptolib"
	"chainhealth/src/cryptolib/tdh2easy"
	monitorLog "chainhealth/src/monitor/log"
	pb "chainhealth/src/proto/proto/communication"
	pb1 "chainhealth/src/proto/proto/serverapis"
	quorum "chainhealth/src/replica/quorum"
	"chainhealth/src/storage"
	utils "chainhealth/src/utils"
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"google.golang.org/grpc"
)

var clientTimer int
var Shares utils.StringIntByteMap

const (
	defaultAPIPort = 12000
)

/*
	 For client to send requests to the replicas.
	 Input
		 rtype: a message type defined in .proto fil
		 t1: a timestamp
		 op: operations (serialized json object of a ClientRequest)
		 address: receiver's address

	 Reuse quorum for replicas to check conditions for clients
	 	PP is used to check whether sufficient matching replies have been received
	 	CM is used to check failures. CM is mainly used for client api.
*/
func SendRequest(rtype pb.MessageType, t1 int64, op []byte, address string) {

	// Set up a connection to the server.
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(clientTimer)*time.Millisecond)
	defer cancel()

	nid := consts.FetchReplicaID(address)
	v, err := utils.StringToInt64(nid)

	conn, err := grpc.DialContext(ctx, address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		p := fmt.Sprintf("[Client Sender Error] did not connect to node %s, set it to notlive: %v", nid, err)
		monitorLog.PrintLog(verbose, monitorLog.ErrorLog, p)
		wg.Done()
		return
	}
	defer conn.Close()
	c := pb.NewSendClient(conn)

	r, err := c.SendRequest(ctx, &pb.Request{Type: rtype, Version: consts.FetchVersionNumber(), Request: op})

	if err != nil {
		p := fmt.Sprintf("[Client Sender Error] could not get reply from node %s, set it to notlive: %v", nid, err)
		monitorLog.PrintLog(verbose, monitorLog.ErrorLog, p)
		wg.Done()
		return
	}

	//handle response
	var re string
	var replymsg message.ClientReply
	rtmp := message.DeserializeMessageWithSignature(r.GetMsg())
	replymsg = message.DeserializeClientReply(rtmp.Msg)

	if consts.EvalMode() == 0 && !cryptolib.VerifySig(replymsg.Source, rtmp.Version, rtmp.Msg, rtmp.Sig) {
		p := fmt.Sprintf("[Client Sender Error] Signature verification for Clientreply from Replica %d failed", replymsg.Source)
		monitorLog.PrintLog(verbose, monitorLog.ErrorLog, p)
		wg.Done()
		return
	}

	if storage.IsCompleted(string(op)) {
		wg.Done()
		return
	}
	p := fmt.Sprintf("[Client Sender] Add the Reply from Replica %s to \"PP\" buffer", nid)
	monitorLog.PrintLog(verbose, monitorLog.NormalLog, p)

	re = utils.BytesToString(replymsg.Msg)

	quorum.Add(v, re, []byte(""), quorum.PP, 0)
	if replymsg.Share != nil {
		v_int, _ := utils.Int64ToInt(v)
		Shares.Insert(re, v_int, replymsg.Share)
	}

	if quorum.CheckSmallQuorum(re) {
		t2 := utils.MakeTimestamp()
		log.Printf("[Reply Info] %s", replymsg.Re)
		p = fmt.Sprintf("[Reply Info] %s", replymsg.Re)
		monitorLog.PrintLog(verbose, monitorLog.NormalLog, p)
		if replymsg.Result == false {
			storage.InsertReturnInvalid(utils.BytesToString(op), utils.StringToBytes(replymsg.Re))
		} else {
			switch replymsg.Type {
			case pb.MessageType_READ:
				//ops := message.DeserializeReadReply(replymsg.Msg)
				//for i := 0; i < len(ops.OPS); i++ {
				//	wr := message.DeserializeWriteRequest(ops.OPS[i].OP)
				//	log.Printf("[Reply] OP %s", wr.OP)
				//}
				s, exist := Shares.Get(re)
				if exist {
					received_shares_value := s.GetAll()
					h, _ := hex.DecodeString(replymsg.Re)
					diagnosis, success := AggregateShares(received_shares_value, utils.StringToBytes(re), h)
					if success {
						var cipherResponse = message.CipherResponse{
							C:      utils.StringToBytes(re),
							Hash:   h,
							Shares: received_shares_value,
						}
						cipher, _ := cipherResponse.Serialize()
						storage.InsertReturnValid(utils.BytesToString(op), diagnosis)
						storage.InsertCipherInfo(utils.BytesToString(op), cipher)
					}
				} else {
					storage.InsertReturnValid(utils.BytesToString(op), replymsg.Msg)
				}
			default:
				storage.InsertReturnValid(utils.BytesToString(op), replymsg.Msg)
			}
		}

		log.Printf("[Client Sender] Latency %d ms", t2-t1)
		p = fmt.Sprintf("[Client Sender] Latency %d ms", t2-t1)
		monitorLog.PrintLog(verbose, monitorLog.NormalLog, p)
		Shares.Delete(re)
		quorum.ClearBuffer(re, quorum.PP)
		storage.Complete(utils.BytesToString(op))
	}
	wg.Done()
}

func SendLoadRequest(rtype pb.MessageType, t1 int64, op []byte, address string) {

	// Set up a connection to the server.
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(clientTimer)*time.Millisecond)
	defer cancel()

	nid := consts.FetchReplicaID(address)

	conn, err := grpc.DialContext(ctx, address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		p := fmt.Sprintf("[Client Sender Error] did not connect to node %s: %v", nid, err)
		monitorLog.PrintLog(verbose, monitorLog.ErrorLog, p)
		return
	}
	defer conn.Close()
	c := pb.NewSendClient(conn)

	_, err = c.SendRequest(ctx, &pb.Request{Type: rtype, Version: consts.FetchVersionNumber(), Request: op})

	if err != nil {
		p := fmt.Sprintf("[Client Sender Error] could not get reply from node %s, %v", nid, err)
		monitorLog.PrintLog(verbose, monitorLog.ErrorLog, p)
		return
	}
}

/*
Broadcast a client request to all other replicas.
*/
func BroadcastRequest(rtype pb.MessageType, op []byte) {
	t1 := utils.MakeTimestamp()

	nodes := FetchNodesFromConfig()

	for i := 0; i < len(nodes); i++ {
		nid := nodes[i]
		if communication.IsNotLive(nid) {
			p := fmt.Sprintf("[Client Sender] Replica %s is not live, not sending any message to it", nid)
			monitorLog.PrintLog(verbose, monitorLog.NormalLog, p)
			continue
		}
		wg.Add(1)
		p := fmt.Sprintf("[Client Sender] Send a %v Request to Replica %v", rtype, nid)
		monitorLog.PrintLog(verbose, monitorLog.NormalLog, p)
		go SendRequest(rtype, t1, op, consts.FetchAddress(nid))
	}
	wg.Wait()
}

/*
Broadcast a client request to all other replicas.
*/
func BroadcastLoadRequest(rtype pb.MessageType, op []byte) {
	t1 := utils.MakeTimestamp()

	nodes := FetchNodesFromConfig()

	for i := 0; i < len(nodes); i++ {
		nid := nodes[i]
		if communication.IsNotLive(nid) {
			p := fmt.Sprintf("[Client Sender] Replica %s is not live, not sending any message to it", nid)
			monitorLog.PrintLog(verbose, monitorLog.NormalLog, p)
			continue
		}
		p := fmt.Sprintf("[Client Sender] Send a %v Request to Replica %v", rtype, nid)
		monitorLog.PrintLog(verbose, monitorLog.NormalLog, p)
		SendLoadRequest(rtype, t1, op, consts.FetchAddress(nid))
	}
}

func BroadcastServerAPI(nid string, sseq int64, eseq int64, t message.TypeOfServerAPI) ([]byte, bool) {
	var group sync.WaitGroup
	nodes := consts.FetchNodes()
	rep := make(chan []byte)
	result := make(chan bool)

	group.Add(len(nodes))
	for i := 0; i < len(nodes); i++ {
		rid := nodes[i]
		if communication.IsNotLive(rid) {
			fmt.Printf("[Client Sender] Replica %s is not live, not sending any message to it \n", rid)
			group.Done()
			continue
		}
		switch t {
		case message.GetChainInfo:
			go SendGetChainInfo(rid, rep, result, &group)
		case message.GetBlockInfo:
			go SendGetLatestBlock(rid, sseq, eseq, rep, result, &group)
		case message.GetNodeInfo:
			go SendGetNodeInfo(nid, rep, result, &group)
		case message.GetTDH2PublicKey:
			go SendGetTDH2PublicKey(nid, rep, result, &group)

		}

		r := <-rep
		success := <-result
		if success != true {
			fmt.Printf("[Client Sender Error] could not get rep from node %s", rid)
			return r, success
		}
		ridint, _ := utils.StringToInt64(rid)
		quorum.Add(ridint, string(r), []byte(""), quorum.PP, 0)
		qresult := quorum.CheckSmallQuorum(string(r))
		if qresult {
			quorum.ClearBuffer(string(r), quorum.PP)
			return r, success
		}
	}

	close(rep)
	close(result)
	group.Wait()
	return nil, false
}

func SendGetNodeInfo(rid string, reply chan<- []byte, result chan<- bool, group *sync.WaitGroup) {
	ridInt, _ := utils.StringToInt(rid)
	address := consts.FetchNodeHost(rid) + ":" + utils.IntToString(defaultAPIPort+ridInt)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(consts.FetchClientTimer())*time.Millisecond)
	defer cancel()
	conn, err := grpc.DialContext(ctx, address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		p := fmt.Sprintf("[Client Sender Error] did not connect to node %s: %v", rid, err)
		monitorLog.PrintLog(verbose, monitorLog.ErrorLog, p)
		group.Done()
		return
	}
	defer conn.Close()
	c := pb1.NewSendClient(conn)
	if err != nil {
		p := fmt.Sprintf("[Client SendRequest Error] getSendClient falied error: %v", err)
		monitorLog.PrintLog(verbose, monitorLog.ErrorLog, p)
		group.Done()
		return
	}

	ridint, _ := utils.StringToInt64(rid)
	r, err := c.GetStaticsOfNode(ctx, &pb1.Input{Nid: ridint})
	if err != nil {
		log.Printf("[ServerAPI ObtainAllData] could not create the transaction: %v", err)
	}
	reply <- r.GetValue()
	result <- r.GetResult()
	group.Done()
}

func SendGetChainInfo(rid string, reply chan<- []byte, result chan<- bool, group *sync.WaitGroup) {
	ridInt, _ := utils.StringToInt(rid)
	address := consts.FetchNodeHost(rid) + ":" + utils.IntToString(defaultAPIPort+ridInt)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(consts.FetchClientTimer())*time.Millisecond)
	defer cancel()

	conn, err := grpc.DialContext(ctx, address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		p := fmt.Sprintf("[Client Sender Error] did not connect to node %s: %v", rid, err)
		monitorLog.PrintLog(verbose, monitorLog.ErrorLog, p)
		group.Done()
		return
	}
	defer conn.Close()
	c := pb1.NewSendClient(conn)
	if err != nil {
		p := fmt.Sprintf("[Client SendRequest Error] getSendClient falied error: %v", err)
		monitorLog.PrintLog(verbose, monitorLog.ErrorLog, p)
		group.Done()
		return
	}

	ridint, _ := utils.StringToInt64(rid)
	r, err := c.GetStaticsOfChain(ctx, &pb1.Input{Nid: ridint})
	if err != nil {
		log.Printf("[ServerAPI ObtainAllData] could not create the transaction: %v", err)
	}
	reply <- r.GetValue()
	result <- r.GetResult()
	group.Done()
}

func SendGetLatestBlock(rid string, sseq int64, eseq int64, reply chan<- []byte, result chan<- bool, group *sync.WaitGroup) {
	ridInt, _ := utils.StringToInt(rid)
	address := consts.FetchNodeHost(rid) + ":" + utils.IntToString(defaultAPIPort+ridInt)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(consts.FetchClientTimer())*time.Millisecond)
	defer cancel()
	conn, err := grpc.DialContext(ctx, address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		p := fmt.Sprintf("[Client Sender Error] did not connect to node %s: %v", rid, err)
		monitorLog.PrintLog(verbose, monitorLog.ErrorLog, p)
		group.Done()
		return
	}
	defer conn.Close()
	c := pb1.NewSendClient(conn)

	ridint, _ := utils.StringToInt64(rid)
	r, err := c.GetLatestBlock(ctx, &pb1.Input{Nid: ridint, Sseq: sseq, Eseq: eseq})
	if err != nil {
		log.Printf("[ServerAPI ObtainAllData] could not create the transaction: %v", err)
	}
	reply <- r.GetValue()
	result <- r.GetResult()
	group.Done()
}

func SendGetTDH2PublicKey(rid string, reply chan<- []byte, result chan<- bool, group *sync.WaitGroup) {
	ridInt, _ := utils.StringToInt(rid)
	address := consts.FetchNodeHost(rid) + ":" + utils.IntToString(defaultAPIPort+ridInt)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(consts.FetchClientTimer())*time.Millisecond)
	defer cancel()
	conn, err := grpc.DialContext(ctx, address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		p := fmt.Sprintf("[Client Sender Error] did not connect to node %s: %v", rid, err)
		monitorLog.PrintLog(verbose, monitorLog.ErrorLog, p)
		group.Done()
		return
	}
	defer conn.Close()
	c := pb1.NewSendClient(conn)
	if err != nil {
		p := fmt.Sprintf("[Client SendRequest Error] getSendClient falied error: %v", err)
		monitorLog.PrintLog(verbose, monitorLog.ErrorLog, p)
		group.Done()
		return
	}

	ridint, _ := utils.StringToInt64(rid)
	r, err := c.GetTDH2PublicKey(ctx, &pb1.Input{Nid: ridint})
	if err != nil {
		log.Printf("[ServerAPI ObtainAllData] could not create the transaction: %v", err)
	}
	reply <- r.GetValue()
	result <- r.GetResult()
	group.Done()
}

/*
Start client parameters for clients
*/
func StartClientSender(cid string, loadkey bool) {
	verbose = consts.FetchVerbose()

	id, err = utils.StringToInt64(cid) // string to int64
	if err != nil {
		p := fmt.Sprintf("[Client Sender Error] Client id %v is not valid. Double check the configuration file", id)
		monitorLog.PrintLog(verbose, monitorLog.ErrorLog, p)
		return
	}

	listOfNodes, _ := utils.ConvertStrArrToIntArr(consts.FetchNodes())
	if loadkey {
		cryptolib.StartCrypto(id, consts.CryptoOption(), listOfNodes)
	} else {
		cryptolib.StartCrypto(id, consts.CryptoOption(), listOfNodes)
	}

	storage.StartClientStorage(cid)
	quorum.StartQuorum(consts.FetchNumReplicas(), verbose)
	communication.StartConnectionManager()
	Shares.Init()

	clientTimer = consts.FetchClientTimer()
	sleepTimerValue = consts.FetchSleepTimer()

}

func removeNilValuesAndBuildNewSlice(slice []*tdh2easy.DecryptionShare) []*tdh2easy.DecryptionShare {
	var newSlice []*tdh2easy.DecryptionShare
	for _, value := range slice {
		if value != nil {
			newSlice = append(newSlice, value)
		}
	}
	return newSlice
}

func AggregateShares(received_shares_value map[int][]byte, c []byte, hash []byte) ([]byte, bool) {

	consts.LoadConfig(0)
	num_nodes := consts.FetchNumReplicas()

	//read public key
	pk_file, _ := os.ReadFile("./etc/test/pk.pem")

	var publicKey tdh2easy.PublicKey
	if err := publicKey.Unmarshal(pk_file); err != nil {
		fmt.Println("Decoding fail:", err)
	}

	//read shares
	shares := make(map[int]*tdh2easy.DecryptionShare)
	for key, value := range received_shares_value {
		var tmp tdh2easy.DecryptionShare
		if err := tmp.Unmarshal(value); err != nil {
			fmt.Println("Unmarshal decryption share fail:", err)
		}
		shares[key] = &tmp
	}

	/*shares2 := make([]*tdh2easy.DecryptionShare, num_nodes)
	for i := 0; i < num_nodes; i++ {
		share_file, _ := os.ReadFile("./etc/test/share" + strconv.Itoa(i) + ".txt")
		var tmp tdh2easy.DecryptionShare
		if err := tmp.Unmarshal(share_file); err != nil {
			fmt.Println("Decoding fail:", err)
		}
		shares2[i] = &tmp
	}*/

	//read cipher
	var deserialized_c tdh2easy.Ciphertext
	if err := deserialized_c.Unmarshal(c); err != nil {
		fmt.Println("Unmarshal ciphertext fail:", err)
	}

	//Second verify the shares Combine the shares
	var ss = make([]*tdh2easy.DecryptionShare, num_nodes)
	for i, share := range shares {
		err := tdh2easy.VerifyShare(&deserialized_c, &publicKey, share)
		if err != nil {
			return []byte("密文分片验证失败！"), false
		} else {
			ss[i] = share
		}
	}

	newSlice := removeNilValuesAndBuildNewSlice(ss)
	plaintext, _ := tdh2easy.Aggregate(&deserialized_c, newSlice, num_nodes)
	return plaintext, true
}
