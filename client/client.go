package main

import (
	"context"
	"flag"
	"fmt"
	"grpc-tls/tlstool"
	"log"
	"math/rand"
	"strings"
	"time"

	"google.golang.org/grpc"
	pb "grpc-tls/helloworld"
)

var (
	serverList = flag.String("server_list", "node1", "server list")
)

func todoFunc(clientid string, server string, seridx int) {
	basePort := 9000
	port := basePort + seridx

	c, err := tlstool.GetClientTlsConfig(clientid, server)
	if err != nil {
		log.Fatalf("get client tls config err: %v", err)
	}

	// create client connection
	conn, err := grpc.Dial(
		fmt.Sprintf("0.0.0.0:%d", port),
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

	resp, err := client.SayHello(ctx, &pb.HelloRequest{Name: server})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Greeting: %s", resp.GetMessage())

}

func main() {
	flag.Parse()
	sList := strings.Split(*serverList, ",")
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			cidx := rand.Intn(len(sList))
			sidx := rand.Intn(len(sList))
			todoFunc(sList[cidx], sList[sidx], sidx)
		}
	}

}
