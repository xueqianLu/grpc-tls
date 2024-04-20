package main

import (
	"context"
	"flag"
	"fmt"
	"grpc-tls/tlstool"
	"log"
	"net"
	"strings"

	"google.golang.org/grpc"
	pb "grpc-tls/helloworld"
)

type greeterService struct {
	pb.UnimplementedGreeterServiceServer
}

func (s *greeterService) SayHello(ctx context.Context, request *pb.HelloRequest) (*pb.HelloReply, error) {
	log.Printf("received name: %v", request.GetName())
	return &pb.HelloReply{Message: "Hello " + request.GetName()}, nil
}

var (
	serverList = flag.String("server_list", "node1", "server list")
)

func main() {
	flag.Parse()
	basePort := 9000
	sList := strings.Split(*serverList, ",")
	for idx, s := range sList {
		port := basePort + idx

		// 公钥中读取和解析公钥/私钥对
		cred, err := tlstool.GetServerTlsConfig(s)
		if err != nil {
			log.Printf("get server tls for (%s) err: %v", s, err)
			continue
		}
		// create grpc server
		grpcServer := grpc.NewServer(grpc.Creds(cred))
		// listen port
		lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
		if err != nil {
			log.Printf("listen port err: %v", err)
			continue
		}

		// register service into grpc server
		pb.RegisterGreeterServiceServer(grpcServer, &greeterService{})
		log.Printf("server: %v at port %d\n", s, port)

		// listen port
		go func() {
			if err := grpcServer.Serve(lis); err != nil {
				log.Printf("grpc serve err: %v", err)
			}
		}()
	}

	<-make(chan struct{})
}
