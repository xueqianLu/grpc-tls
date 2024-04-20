package main

import (
	"context"
	"grpc-tls/tlstool"
	"log"
	"net"

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

func main() {
	// 公钥中读取和解析公钥/私钥对
	cred, err := tlstool.GetServerTlsConfig("node.chain.app")
	if err != nil {
		log.Fatalf("get server tls config err: %v", err)
	}
	// create grpc server
	grpcServer := grpc.NewServer(grpc.Creds(cred))
	// listen port
	lis, err := net.Listen("tcp", "0.0.0.0:9000")
	if err != nil {
		log.Fatalf("list port err: %v", err)
	}

	// register service into grpc server
	pb.RegisterGreeterServiceServer(grpcServer, &greeterService{})

	log.Printf("listening at %v", lis.Addr())

	// listen port
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("grpc serve err: %v", err)
	}
}
