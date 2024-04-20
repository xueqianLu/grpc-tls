package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	pb "grpc-tls/helloworld"
)

func main() {
	// 公钥中读取和解析公钥/私钥对
	cert, err := tls.LoadX509KeyPair("./conf/client/client.crt", "./conf/client/client.key")
	if err != nil {
		fmt.Println("LoadX509KeyPair error ", err)
		return
	}
	// 创建一组根证书
	certPool := x509.NewCertPool()
	ca, err := ioutil.ReadFile("./conf/ca.crt")
	if err != nil {
		fmt.Println("ReadFile ca.crt error ", err)
		return
	}
	// 解析证书
	if ok := certPool.AppendCertsFromPEM(ca); !ok {
		fmt.Println("certPool.AppendCertsFromPEM error ")
		return
	}

	c := credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{cert},
		ServerName:   "node.chain.app",
		RootCAs:      certPool,
	})
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

	resp, err := client.SayHello(ctx, &pb.HelloRequest{Name: "Mert Kimyonsen"})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Greeting: %s", resp.GetMessage())

}

func main1() {
	// read ca's cert
	caCert, err := ioutil.ReadFile("cert/ca-cert.pem")
	if err != nil {
		log.Fatal(caCert)
	}

	// create cert pool and append ca's cert
	certPool := x509.NewCertPool()
	if ok := certPool.AppendCertsFromPEM(caCert); !ok {
		log.Fatal(err)
	}

	//read client cert
	clientCert, err := tls.LoadX509KeyPair("cert/client-cert.pem", "cert/client-key.pem")
	if err != nil {
		log.Fatal(err)
	}

	config := &tls.Config{
		Certificates: []tls.Certificate{clientCert},
		ServerName:   "0.0.0.0",
		RootCAs:      certPool,
	}

	tlsCredential := credentials.NewTLS(config)

	// create client connection
	conn, err := grpc.Dial(
		"0.0.0.0:9000",
		grpc.WithTransportCredentials(tlsCredential),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := pb.NewGreeterServiceClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	resp, err := client.SayHello(ctx, &pb.HelloRequest{Name: "Mert Kimyonsen"})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Greeting: %s", resp.GetMessage())
}
