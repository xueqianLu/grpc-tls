package communication

import (
	"chainhealth/src/communication/pool"
	pb "chainhealth/src/proto/proto/communication"
	"errors"
	"fmt"
	"log"
	"sync"
)

type Connection struct {
	sync.RWMutex
	pool    map[string]pool.Pool
	Options pool.Options
}

type Communication interface {
	AddConnect(addr string)
	ReBuildConnect(addr string)
	CloseConnect(addr string)
	GetConnect(addr string) (pool.Pool, bool)
}

func (connection *Connection) AddConnect(addr string) {
	connection.Lock()
	defer connection.Unlock()
	p, err := pool.New(addr, connection.Options)
	if err != nil {
		return
	}
	connection.pool[addr] = p
}

func (connection *Connection) ReBuildConnect(addr string) {
	connection.Lock()
	defer connection.Unlock()
	p, err := pool.New(addr, connection.Options)
	if err != nil {
		return
	}
	oldPool, suc := connection.pool[addr]
	if suc {
		oldPool.Close()
	}
	connection.pool[addr] = p
}

func (connection *Connection) CloseConnect(addr string) {
	connection.Lock()
	defer connection.Unlock()
	oldPool, suc := connection.pool[addr]
	if suc {
		oldPool.Close()
	}
	delete(connection.pool, addr)
}

func (connection *Connection) GetConnect(addr string) (pool.Pool, bool) {
	connection.RLock()
	defer connection.RUnlock()
	grpcPool, exist := connection.pool[addr]
	if exist {
		return grpcPool, true
	}
	return nil, false
}

func (connection *Connection) NewSendClient(address string) (pb.SendClient, pool.Conn, error) {
	grpcPool, suc := connection.GetConnect(address)
	if !suc || grpcPool == nil {
		p := fmt.Sprintf("failed to get address %v grpc pool", address)
		return nil, nil, errors.New(p)
	}
	conn, err := grpcPool.Get()
	if err != nil || conn == nil {
		p := fmt.Sprintf("failed to get address %v conn:%v ", address, err)
		return nil, nil, errors.New(p)
	}
	if conn == nil || err != nil {
		p := fmt.Sprintf("Get address %v conn error:%v,conn:%v", address, err, conn)
		log.Println(p)
		return nil, nil, errors.New(p)
	}

	c := pb.NewSendClient(conn.Value())
	return c, conn, nil
}
