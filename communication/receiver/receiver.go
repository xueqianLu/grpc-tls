/*
Receiver functions.
It implements all the gRPC services defined in communication.proto file.
*/

package receiver

import (
	"bytes"
	"chainhealth/src/communication/message"
	"chainhealth/src/consts"
	"chainhealth/src/cryptolib"
	monitorLog "chainhealth/src/monitor/log"
	pb "chainhealth/src/proto/proto/communication"
	"chainhealth/src/replica/handler"
	"chainhealth/src/storage"
	"chainhealth/src/utils"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
	"sync"
)

var id string
var wg sync.WaitGroup

type server struct {
	pb.UnimplementedSendServer
}

/*
Handling view change and new view messages.
Replicas do not have to reply to the VC messages (reply with a pb.Empty message)
*/
func (s *server) SendVCMsg(ctx context.Context, in *pb.RawMessage) (*pb.Empty, error) {
	go handler.HandleVCMsg(in.GetMsg())
	return &pb.Empty{}, nil
}

/*
Handling checkpoint messages.
Replicas do not reply to checkpoint messages (reply with a pb.Empty message)
*/
func (s *server) SendCheckpoint(ctx context.Context, in *pb.RawMessage) (*pb.Empty, error) {
	go handler.HandleCheckpoint(in.GetMsg())
	return &pb.Empty{}, nil
}

/*
Handle replica messages (consensus normal operations)
*/
func (s *server) RBCSendByteMsg(ctx context.Context, in *pb.RawMessage) (*pb.Empty, error) {
	go handler.HandleByteMsg(in.GetMsg())
	return &pb.Empty{}, nil
}

/*
Handle all client requests but different operations are handled in different functions.
Each client expects a reply per a request messages.
Use channel to listen to reply messages and return once a reply is available.
*/
func (s *server) SendRequest(ctx context.Context, in *pb.Request) (*pb.RawMessage, error) {
	h := cryptolib.GenHash(in.GetRequest())
	rtype := in.GetType()

	//用户权限验证
	switch rtype {
	case pb.MessageType_READ, pb.MessageType_WRITE, pb.MessageType_WRITEBATCH:
		re, success := handler.VerifyUserRequest(in.GetRequest())
		if !success {
			return &pb.RawMessage{Msg: handler.GenerateErrorResponse(rtype, re, h)}, nil
		}
	case pb.MessageType_CONTRACT:
		re, success := handler.VerifyUserRequest(in.GetRequest())
		if !success {
			return &pb.RawMessage{Msg: handler.GenerateErrorResponse(rtype, re, h)}, nil
		}
		err := storage.EvmHandler.VerirfyContractTx(in.GetRequest())
		if err != nil {
			return &pb.RawMessage{Msg: handler.GenerateErrorResponse(rtype, err.Error(), h)}, nil
		}
	case pb.MessageType_DYNAMICNODE, pb.MessageType_ADMIN:
		re, success := handler.VerifyAdminRequest(in.GetRequest())
		if !success {
			return &pb.RawMessage{Msg: handler.GenerateErrorResponse(rtype, re, h)}, nil
		}
	}

	switch rtype {
	case pb.MessageType_READ:
		reply := handler.HandleReadRequestStruct(in.GetRequest())
		return &pb.RawMessage{Msg: reply}, nil

	case pb.MessageType_WRITEBATCH:
		msgs, hashes := handler.GetMsgs(in.GetRequest())
		for i := 0; i < len(msgs); i++ {
			go handler.HandleBatchRequest(msgs[i], hashes[i])
		}
		replies := make(chan []byte)
		go handler.GetTempResponseViaChan(replies)
		reply := <-replies
		//reply := handler.GetTempResponse()
		return &pb.RawMessage{Msg: reply}, nil
	}

	go handler.HandleRequest(in.GetRequest(), utils.BytesToString(h))

	if consts.EvalMode() > 0 {
		return &pb.RawMessage{Msg: []byte("Test")}, nil
	}

	replies := make(chan []byte)
	go handler.GetResponseViaChan(rtype, utils.BytesToString(h), replies)
	reply := <-replies

	return &pb.RawMessage{Msg: reply}, nil
}

/*
Handle join requests for both static membership (initialization) and dynamic membership.
Each replica gets a conformation for a membership request.
*/
func (s *server) Join(ctx context.Context, in *pb.RawMessage) (*pb.RawMessage, error) {
	reply := handler.HandleJoinRequest(in.GetMsg())
	result := true
	if bytes.Compare(reply, []byte("")) == 0 {
		result = false
	}
	return &pb.RawMessage{Msg: reply, Result: result}, nil
}

/*
Handle catchup request for both static membership (initializatoin) and dynamic membership.
*/
func (s *server) SendCatchUp(ctx context.Context, in *pb.RawMessage) (*pb.Empty, error) {
	handler.HandleStateTransfer(in.GetMsg())
	return &pb.Empty{}, nil
}

func (s *server) LargeStateTransfer(ctx context.Context, in *pb.RawMessage) (*pb.Empty, error) {
	handler.HandleLargeStateTransfer(in.GetMsg())
	return &pb.Empty{}, nil
}

func (s *server) Sync(ctx context.Context, in *pb.RawMessage) (*pb.RawMessage, error) {
	reply, result := handler.HandleSyncRequest(in.GetMsg())
	return &pb.RawMessage{Msg: reply, Result: result}, nil
}

/*
Register rpc socket via port number and ip address
*/
func register(port string) {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		p := fmt.Sprintf("[Communication Receiver Error] failed to listen %v", err)
		monitorLog.PrintLog(consts.FetchVerbose(), monitorLog.ErrorLog, p)
		os.Exit(1)
	}

	p := fmt.Sprintf("[Communication Receiver] listening to port %v", port)
	monitorLog.PrintLog(consts.FetchVerbose(), monitorLog.NormalLog, p)

	defer wg.Done()
	s := grpc.NewServer(grpc.MaxRecvMsgSize(52428800), grpc.MaxSendMsgSize(52428800))
	pb.RegisterSendServer(s, &server{})

	if err := s.Serve(lis); err != nil {
		p := fmt.Sprintf("[Communication Receiver Error] failed to serve: %v", err)
		monitorLog.PrintLog(consts.FetchVerbose(), monitorLog.ErrorLog, p)
		os.Exit(1)
	}

}

/*
Start receiver parameters initialization
*/
func StartReceiver(rid string, joinOption int, dest string, count int) {
	id = rid
	monitorLog.SetLog(rid)
	consts.LoadConfig(0)

	memtype := message.TypeOfMembership_name[joinOption]

	if !consts.CleanDB() && memtype == message.ExiRepCleanDB {
		monitorLog.PrintLog(consts.FetchVerbose(), monitorLog.ErrorLog, "[Receiver Error] DB has not been cleared but the replica joins the system as a node with clean DB. Data might become inconsistent")
		log.Printf("[Receiver Error] DB has not been cleared but the replica joins the system as a node with clean DB. Data might become inconsistent")
		os.Exit(1)
	}

	if consts.CleanDB() && memtype == message.ExiRepDB {
		monitorLog.PrintLog(consts.FetchVerbose(), monitorLog.ErrorLog, "[Receiver WARNING] DB has been cleared but the replica joins the system as a node with DB data. Data after last stable checkpoint will be obtained.")
		log.Printf("[Receiver WARNING] DB has been cleared but the replica joins the system as a node with DB data. Data after last stable checkpoint will be obtained.")
		os.Exit(1)
	}

	if !consts.ExistsInConfig(rid) {
		p := fmt.Sprintf("[Receiver Error] Node %s does not exist in configuration file, cannot join the system.", rid)
		monitorLog.PrintLog(consts.FetchVerbose(), monitorLog.ErrorLog, p)
		log.Printf("[Receiver Error] Node %s does not exist in configuration file, cannot join the system.", rid)
		os.Exit(1)
	}
	handler.StartHandler(rid, joinOption, dest, count)
	wg.Add(1)
	go StartServerAPI(rid)
	register(consts.FetchPort(rid))
	wg.Wait()

}
