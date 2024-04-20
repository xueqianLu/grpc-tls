package receiver

import (
	"chainhealth/src/consts"
	"chainhealth/src/cryptolib"
	monitorLog "chainhealth/src/monitor/log"
	pb1 "chainhealth/src/proto/proto/serverapis"
	"chainhealth/src/replica/handler"
	"chainhealth/src/utils"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
)

const (
	defaultPortNum = 12000
)

type serverAPI struct {
	pb1.UnimplementedSendServer
}

var portnumber string

/*
Get chain information bt serverAPI.
*/
func (s *serverAPI) GetStaticsOfChain(ctx context.Context, in *pb1.Input) (*pb1.Output, error) {
	value := handler.StaticsOfChain()
	data, _ := value.Serialize()
	return &pb1.Output{Value: data, Result: true}, nil
}

func (s *serverAPI) GetStaticsOfNode(ctx context.Context, in *pb1.Input) (*pb1.Output, error) {
	value := handler.StaticsOfNode()
	data, _ := value.Serialize()
	return &pb1.Output{Value: data, Result: true}, nil
}

func (s *serverAPI) GetLatestBlock(ctx context.Context, in *pb1.Input) (*pb1.Output, error) {
	value := handler.LatestBlock(in.Sseq, in.Eseq)
	data, _ := value.Serialize()
	return &pb1.Output{Value: data, Result: true}, nil
}

func (s *serverAPI) GetTDH2PublicKey(ctx context.Context, in *pb1.Input) (*pb1.Output, error) {
	data, result := cryptolib.GetTDH2PublickKey()
	return &pb1.Output{Value: data, Result: result}, nil
}

/*
Register rpc socket via port number and ip address
*/
func registerAPI(port string) {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		p := fmt.Sprintf("[ServerAPI Error] failed to listen %v", err)
		monitorLog.PrintLog(consts.FetchVerbose(), monitorLog.ErrorLog, p)
		os.Exit(1)
	}

	s := grpc.NewServer(grpc.MaxRecvMsgSize(52428800), grpc.MaxSendMsgSize(52428800))
	pb1.RegisterSendServer(s, &serverAPI{})

	log.Printf("[ServerAPI] listening at %v", port)
	if err := s.Serve(lis); err != nil {
		p := fmt.Sprintf("[ServerAPI Receiver Error] failed to serve: %v", err)
		monitorLog.PrintLog(consts.FetchVerbose(), monitorLog.ErrorLog, p)
		os.Exit(1)
	}

}

/*
Get port number for server api
*/
func GetPortNumber(rid string) string {
	portNum := defaultPortNum
	ridInt, _ := utils.StringToInt(rid)
	pn := ":" + utils.IntToString(portNum+ridInt)
	return pn
}

/*
Start receiver to listen to the frontend messages
*/
func StartServerAPI(rid string) {
	portnumber = GetPortNumber(rid)
	registerAPI(portnumber)
}
