package main

import (
	"context"
	"grpc-tls/tlstool"
	"log"
	"time"

	"google.golang.org/grpc"
	pb "grpc-tls/helloworld"
)

func main() {
	c, err := tlstool.GetClientTlsConfig("node1", "node1")
	if err != nil {
		log.Fatalf("get client tls config err: %v", err)
	}

	// create client connection
	conn, err := grpc.Dial(
		"0.0.0.0:9000",
		grpc.WithTransportCredentials(c),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := pb.NewGreeterServiceClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	resp, err := client.SayHello(ctx, &pb.HelloRequest{Name: "Sager"})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Greeting: %s", resp.GetMessage())
}
