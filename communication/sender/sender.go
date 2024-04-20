/*
Sender functions.
It implements all sending functions for replicas.
*/

package sender

import (
	"chainhealth/src/communication"
	"chainhealth/src/communication/message"
	"chainhealth/src/consts"
	monitorLog "chainhealth/src/monitor/log"
	pb "chainhealth/src/proto/proto/communication"
	quorum "chainhealth/src/replica/quorum"
	"chainhealth/src/utils"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"os"
	"strconv"
	"sync"
	"time"
)

var id int64
var err error
var verbose bool

var wg sync.WaitGroup

var broadcastTimer int
var sleepTimerValue int
var reply []byte
var leaveRequest bool

/*
Send state transfer request
*/
func SendStateTransfer(msg []byte, address string, nodeid string) ([]byte, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Duration(broadcastTimer)*time.Millisecond)
	defer cancel()

	conn, err := grpc.DialContext(ctx, address, grpc.WithInsecure(), grpc.WithBlock())

	if err != nil {
		return []byte(""), false
	}

	defer conn.Close()
	c := pb.NewSendClient(conn)

	r, errr := c.Sync(ctx, &pb.RawMessage{Msg: msg, Version: consts.FetchVersionNumber()})
	if errr != nil {
		p := fmt.Sprintf("[Communication Sender Error] Join Error: could not get reply from node %s: %v", nodeid, errr)
		monitorLog.PrintLog(verbose, monitorLog.ErrorLog, p)
		return []byte(""), false
	}
	if consts.TypeOfError_name[string(r.GetMsg())] == consts.VersionError {
		p := fmt.Sprintf("[Communication Sender Error] Join Error: version error from node %s", nodeid)
		monitorLog.PrintLog(verbose, monitorLog.ErrorLog, p)
		return []byte(""), false
	}

	return r.GetMsg(), r.GetResult()
}

func SendLargeState(msg []byte, nodeid int64) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Duration(broadcastTimer)*time.Millisecond)
	defer cancel()

	address := consts.FetchAddress(utils.Int64ToString(nodeid))
	conn, err := grpc.DialContext(ctx, address, grpc.WithInsecure(), grpc.WithBlock())

	if err != nil {
		return
	}

	defer conn.Close()
	c := pb.NewSendClient(conn)

	_, errr := c.LargeStateTransfer(ctx, &pb.RawMessage{Msg: msg})
	if errr != nil {
		p := fmt.Sprintf("[Communication Sender Error] Join Error: could not get reply from node %v: %v", nodeid, errr)
		monitorLog.PrintLog(verbose, monitorLog.ErrorLog, p)
		return
	}

	return
}

/*
Send a join message to the servers. The message is a serialized json format in bytes
Join of a new replica requries each replica to consistently monitor the connections until sufficient number of connections are built.
Therefore, this function is implemented differently from ByteSend.

Input

	msg: serialized join message that needs to be sent ([]byte type)
	address: address of the receiver (ip + port, string type)
	nid: replica id (int type)
*/
func JoinSend(msg []byte, address string, nid int, nodeid string) {
	for {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Duration(consts.FetchBroadcastTimer())*time.Millisecond)
		defer cancel()

		conn, err := grpc.DialContext(ctx, address, grpc.WithInsecure(), grpc.WithBlock())

		if err != nil {
			//fmt.Println(err)
			time.Sleep(time.Duration(sleepTimerValue) * time.Millisecond)
			continue
		}

		defer conn.Close()
		c := pb.NewSendClient(conn)

		r, errr := c.Join(ctx, &pb.RawMessage{Msg: msg, Version: consts.FetchVersionNumber()})
		if errr != nil {
			p := fmt.Sprintf("[Communication Sender Error] Join Error: could not get reply from node %s: %v", nodeid, errr)
			monitorLog.PrintLog(verbose, monitorLog.ErrorLog, p)
			return
		} else {
			if consts.TypeOfError_name[string(r.GetMsg())] == consts.VersionError {
				p := fmt.Sprintf("[Communication Sender Error] Join Error: version error from node %s", nodeid)
				monitorLog.PrintLog(verbose, monitorLog.ErrorLog, p)
				return
			}
			if !leaveRequest {
				p := fmt.Sprintf("[Communication Sender] Connection to node %s established", nodeid)
				monitorLog.PrintLog(verbose, monitorLog.NormalLog, p)
				if !r.GetResult() {
					log.Printf("Join failed.")
					os.Exit(0)
					return
				}
				quorum.AddJoin(nid, r.GetMsg())
				return
			} else {
				quorum.AddMem(nid)
				if quorum.GetMemCount() >= quorum.QuorumSize() {
					log.Printf("Ready to exit.")
					os.Exit(0)
					wg.Done()
				}
				return
			}
		}

	}

}

/*
Send messages in bytes to replicas/clients.
The messages could be ReplicaMessage, ViewChangeMessage, CheckpointMessage, or JoinMsg.

Input

	msg: serialized message ([]byte type)
	address: address of the receiver (ip + port, string type)
	msgType: type of the message (defined in servermessage.go)
*/
func ByteSend(msg []byte, address string, msgType message.TypeOfMessage) {
	// Set up a connection to the server.

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(broadcastTimer)*time.Millisecond)
	defer cancel()

	if address == "" {
		return
	}

	nid := consts.FetchReplicaID(address)

	conn, err := grpc.DialContext(ctx, address, grpc.WithInsecure(), grpc.WithBlock())

	if err != nil {
		p := fmt.Sprintf("[Communication Sender Error] did not connect to node %s: %v", nid, err)
		monitorLog.PrintLog(verbose, monitorLog.ErrorLog, p)
		return
	}

	defer conn.Close()
	c := pb.NewSendClient(conn)

	switch msgType {
	case message.ReplicaMsg:
		_, err = c.RBCSendByteMsg(ctx, &pb.RawMessage{Msg: msg})
		if err != nil {
			p := fmt.Sprintf("[Communication Sender Error] could not get reply from node %s when send ReplicaMsg, set it to notlive: %v", nid, err)
			monitorLog.PrintLog(verbose, monitorLog.ErrorLog, p)
			return
		}
	case message.VCMsg:
		_, err = c.SendVCMsg(ctx, &pb.RawMessage{Msg: msg})
		if err != nil {
			p := fmt.Sprintf("[Communication Sender Error] could not get reply from node %s when send VCMsg, set it to notlive: %v", nid, err)
			monitorLog.PrintLog(verbose, monitorLog.ErrorLog, p)
			return
		}
	case message.CheckpointMsg:
		_, err = c.SendCheckpoint(ctx, &pb.RawMessage{Msg: msg})
		if err != nil {
			p := fmt.Sprintf("[Communication Sender Error] could not get reply from node %s when send CheckpointMsg, set it to notlive: %v", nid, err)
			monitorLog.PrintLog(verbose, monitorLog.ErrorLog, p)
			return
		}
	case message.JoinMsg:
		_, err = c.SendCatchUp(ctx, &pb.RawMessage{Msg: msg})
		if err != nil {
			p := fmt.Sprintf("[Replica Error] could not get reply from node %s: %v", nid, err)
			monitorLog.PrintLog(verbose, monitorLog.ErrorLog, p)
			return
		}
	}

}

/*
Send a message (already signed and []byte format) to a node by node id.
Used for dynamic membership normal operation state transfer. Replicas will perform state transfer with the new node if any.
Input

	msg: serialized message of a CatchUp struct ([]byte type)
	nid: replica id (int type)
*/
func SendMsgToNode(msg []byte, nid int) {
	log.Printf("sending to %v", consts.FetchAddress(utils.IntToString(nid)))
	ByteSend(msg, consts.FetchAddress(utils.IntToString(nid)), message.JoinMsg)
}

/*
Send a message (already signed and []byte format) to a node by host name and port number.
Used for dynamic membership normal operation. Replicas send consensus messages to temporary members.
Input

	msg: serialized message of ReplicaMessage type ([]byte type)
	host: host name (string type)
	port: port number (string type)
*/
func SendByteToNode(msg []byte, host string, port string) {
	addr := host + port
	ByteSend(msg, addr, message.ReplicaMsg)
}

/*
Send a message (already signed and []byte format) to an address (host+ip).
Used for dynamic membership. Replicas send view changes messages to replicas in different configurations
Input

	msg: serialized message of ReplicaMessage type ([]byte type)
	addr: host+ip in the grpc format
*/
func SendByteToAddr(msg []byte, addr string) {
	ByteSend(msg, addr, message.ReplicaMsg)
}

/*
Broadcast a ReplicaMessage in bytes to all other replicas.
Can be optimized to send the messages in parallel.
Input

	msg: serialized messages of any types which will be treated differently ([]byte type)
*/
func RBCByteBroadcast(msg []byte) {
	request, err := message.SerializeWithSignature(id, msg)
	if err != nil {
		monitorLog.PrintLog(verbose, monitorLog.ErrorLog, "[Sender Error] Not able to sign the message")
		return
	}
	nodes := FetchNodesFromConfig()
	for i := 0; i < len(nodes); i++ {
		nid := nodes[i]
		if communication.IsNotLive(nid) {
			p := fmt.Sprintf("[Communication Sender] Replica %v is not live, don't send message to it", i)
			monitorLog.PrintLog(verbose, monitorLog.NormalLog, p)
			continue
		}
		//p := fmt.Sprintf("[Communication Sender] Send a ReplicaMsg to Replica %d",i)
		//monitorLog.PrintLog(verbose, monitorLog.NormalLog,p)
		ByteSend(request, consts.FetchAddress(nid), message.ReplicaMsg)
	}
}

/*
Directly broadcast a message to all replicas.
Used in dynamic membership for normal operation
Input

	msg: serialized message ([]byte format)
*/
func SimpleBroadcast(msg []byte) {
	nodes := FetchNodesFromConfig()
	for i := 0; i < len(nodes); i++ {
		nid := nodes[i]
		if communication.IsNotLive(nid) {
			continue
		}
		ByteSend(msg, consts.FetchAddress(nid), message.ReplicaMsg)
	}
}

/*
Broadcast a checkpoint message in bytes to all other replicas.
Can be optimized to send the messages in parallel.
Input

	msg: serialized checkpoint message ([]byte type)
*/
func SendCheckpoint(msg []byte) {

	nodes := FetchNodesFromConfig()
	for i := 0; i < len(nodes); i++ {
		nid := nodes[i]
		if communication.IsNotLive(nid) {
			p := fmt.Sprintf("[Communication Sender] Replica %v is not live, don,t send the checkpoint message to it", nid)
			monitorLog.PrintLog(verbose, monitorLog.NormalLog, p)
			continue
		}
		p := fmt.Sprintf("[Communication Sender] Send a CheckpointMsg to Replica %v", nid)
		monitorLog.PrintLog(verbose, monitorLog.NormalLog, p)
		ByteSend(msg, consts.FetchAddress(nid), message.CheckpointMsg)
	}
}

/*
Broadcast a view change message in bytes to all other replicas.
Can be optimized to send the messages in parallel.
Input

	msg: serialized view change and new view messages ([]byte type)
*/
func RBCVCBroadcast(msg []byte) {

	request, err := message.SerializeWithSignature(id, msg)
	if err != nil {
		p := fmt.Sprintf("[Sender Error]Not able to sign the message: %v", err)
		monitorLog.PrintLog(verbose, monitorLog.ErrorLog, p)
		return
	}

	nodes := FetchNodesFromConfig()
	for i := 0; i < len(nodes); i++ {
		nid := nodes[i]
		if communication.IsNotLive(nid) {
			p := fmt.Sprintf("[Communication Sender] Replica %v is not live, don,t send the checkpoint message to it", i)
			monitorLog.PrintLog(verbose, monitorLog.NormalLog, p)
			continue
		}
		p := fmt.Sprintf("[Communication Sender] Send a VCMsg to Replica %v", nid)
		monitorLog.PrintLog(verbose, monitorLog.NormalLog, p)
		go ByteSend(request, consts.FetchAddress(nid), message.VCMsg)
	}
}

/*
Broadcast a join message in bytes to all other replicas.
Can be optimized to send the messages in parallel.
Input

	msg: serialized join request ([]byte type)
*/
func SendJoinRequest(msg []byte) {
	nodes := FetchNodesFromConfig()
	for i := 0; i < len(nodes); i++ {
		nid := nodes[i]
		tmp, e := utils.Int64ToInt(id)
		if e != nil || i == tmp {
			continue
		}
		//fmt.Printf("[Communication Sender] Send a Join message to Replica %v\n",nid)
		p := fmt.Sprintf("[Communication Sender] Send a Join message to Replica %v", nid)
		monitorLog.PrintLog(verbose, monitorLog.NormalLog, p)
		go JoinSend(msg, consts.FetchAddress(nid), i, nid)
	}
}

/*
Synchronize with one node
*/
func SendStateTransferRequest(msg []byte, nid string) ([]byte, bool) {
	return SendStateTransfer(msg, consts.FetchAddress(nid), nid)
}

/*
Used for dynamic membership. Broadcast a leave request.
Input

	msg: serialized leave request ([]byte type)
*/
func SendLeaveRequest(msg []byte) {

	nodes := FetchNodesFromConfig()
	for i := 0; i < len(nodes); i++ {
		nid := nodes[i]
		if communication.IsNotLive(nid) {
			continue
		}
		wg.Add(1)

		go JoinSend(msg, consts.FetchAddress(nid), i, nid)
	}
	wg.Wait()
}

/*
Used for membership protocol to fetch address of a node
Input

	rid: id of the node (string type)

Output

	address: address of the node (host+ip, string type)
*/
func FetchAddressFromConfig(rid string) string {
	return consts.FetchAddress(rid)
}

/*
Used for membership protocol to fetch port of a node
Input

	rid: id of the node (string type)

Output

	port: port number (string type)
*/
func FetchPortFromConfig(rid string) string {
	return consts.FetchPort(rid)
}

/*
Used for membership protocol to fetch list of nodes
Output

	[]string: a list of nodes (in the string type)
*/
func FetchNodesFromConfig() []string {
	return consts.FetchNodes()
}

/*
Reload the configuration file
*/
func ReloadConfig() {
	consts.LoadConfig(0)
}

/*
Set leave type to true. Used for dynamic membership only.
*/
func SetLeaveType() {
	leaveRequest = true
}

/*
Start sender parameters for replicas
*/
func StartSender(rid string) {

	monitorLog.PrintLog(verbose, monitorLog.NormalLog, "[Communication Sender] Loading consts for replica")
	consts.LoadConfig(0)
	verbose = consts.FetchVerbose()

	id, err = strconv.ParseInt(rid, 10, 64) // string to int64
	if err != nil {
		p := fmt.Sprintf("[Communication Sender Error] Replica id %v is not valid. Double check the configuration file", id)
		monitorLog.PrintLog(verbose, monitorLog.ErrorLog, p)
		return
	}
	//consts.LoadConfig(0)

	leaveRequest = false
	verbose = consts.FetchVerbose()
	quorum.StartQuorum(consts.FetchNumReplicas(), verbose)
	communication.StartConnectionManager()
	broadcastTimer = consts.FetchBroadcastTimer()
}
