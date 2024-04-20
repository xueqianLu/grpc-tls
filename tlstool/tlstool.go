package tlstool

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"google.golang.org/grpc/credentials"
	"os"
	"path/filepath"
)

func GetClientTlsConfig(clientId string, serverId string) (credentials.TransportCredentials, error) {
	confDir := "./conf"
	if os.Getenv("TLS_CONFIG_DIR") != "" {
		confDir = os.Getenv("TLS_CONFIG_DIR")
	}

	caDir := filepath.Join(confDir, "ca")
	clientDir := filepath.Join(confDir, clientId)

	caCrtPath := filepath.Join(caDir, "ca.crt")
	clientCrtPath := filepath.Join(clientDir, "client.crt")
	clientKeyPath := filepath.Join(clientDir, "client.key")

	// 公钥中读取和解析公钥/私钥对
	cert, err := tls.LoadX509KeyPair(clientCrtPath, clientKeyPath)
	if err != nil {
		fmt.Println("LoadX509KeyPair error ", err)
		return nil, err
	}
	// 创建一组根证书
	certPool := x509.NewCertPool()
	ca, err := os.ReadFile(caCrtPath)
	if err != nil {
		fmt.Println("ReadFile ca.crt error ", err)
		return nil, err
	}
	// 解析证书
	if ok := certPool.AppendCertsFromPEM(ca); !ok {
		fmt.Println("certPool.AppendCertsFromPEM error ")
		return nil, err
	}

	c := credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{cert},
		ServerName:   serverId,
		RootCAs:      certPool,
	})
	return c, nil
}

func GetServerTlsConfig(serverId string) (credentials.TransportCredentials, error) {
	confDir := "./conf"
	if os.Getenv("TLS_CONFIG_DIR") != "" {
		confDir = os.Getenv("TLS_CONFIG_DIR")
	}

	caDir := filepath.Join(confDir, "ca")
	serverDir := filepath.Join(confDir, serverId)

	caCrtPath := filepath.Join(caDir, "ca.crt")
	serverCrtPath := filepath.Join(serverDir, "server.crt")
	serverKeyPath := filepath.Join(serverDir, "server.key")

	// 公钥中读取和解析公钥/私钥对
	cert, err := tls.LoadX509KeyPair(serverCrtPath, serverKeyPath)
	if err != nil {
		fmt.Println("LoadX509KeyPair error ", err)
		return nil, err
	}
	// 创建一组根证书
	certPool := x509.NewCertPool()
	ca, err := os.ReadFile(caCrtPath)
	if err != nil {
		fmt.Println("ReadFile ca.crt error ", err)
		return nil, err
	}
	// 解析证书
	if ok := certPool.AppendCertsFromPEM(ca); !ok {
		fmt.Println("certPool.AppendCertsFromPEM error ")
		return nil, err
	}

	c := credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{cert},
		//要求必须校验客户端的证书
		ClientAuth: tls.RequireAndVerifyClientCert,
		//设置根证书的集合，校验方式使用ClientAuth设定的模式
		ClientCAs: certPool,
	})
	return c, nil
}
