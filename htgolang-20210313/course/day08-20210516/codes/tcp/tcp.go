package main

import (
	"fmt"
	"sync"
)

type TCPConnect struct {
	Addr string
}

type TCPPool struct {
	addr    string
	pool    *sync.Pool
	max     int // 最大数量报错
	idle    int // 池子中保持的最大连接数
	min     int // 池子中最小的连接数
	poolCnt int // 池子连接数量
	count   int // 总的连接数量
	locker  *sync.Mutex
}

func New(addr string, min, max, idle int) *TCPPool {
	pool := &TCPPool{
		min:  min,
		max:  max,
		idle: idle,
		pool: &sync.Pool{
			New: func() interface{} {
				fmt.Println("连接", addr)
				return &TCPConnect{Addr: addr}
			},
		},
		poolCnt: 0,
		count:   0,
		locker:  &sync.Mutex{},
	}
	// 需要初始化连接池数量
	conns := make([]*TCPConnect, 0, min)
	for min > 0 {
		conn, _ := pool.Get()
		conns = append(conns, conn)
		min -= 1
	}

	for _, conn := range conns {
		pool.Put(conn)
	}
	return pool
}

func (p *TCPPool) Get() (*TCPConnect, error) {
	p.locker.Lock()
	defer p.locker.Unlock()
	// 先看poolCnt == 0
	if p.poolCnt > 0 {
		conn := p.pool.Get()
		p.poolCnt -= 1
		return conn.(*TCPConnect), nil
	}

	if p.count >= p.max {
		return nil, fmt.Errorf("max connection")
	}
	p.count += 1
	conn := p.pool.Get()
	return conn.(*TCPConnect), nil

}

func (p *TCPPool) Put(conn *TCPConnect) {
	p.locker.Lock()
	defer p.locker.Unlock()
	if p.poolCnt >= p.idle {
		p.count -= 1 // 关闭连接
		fmt.Println("连接关闭")
		return
	}
	p.poolCnt += 1
	p.pool.Put(conn)
}

func (p *TCPPool) Close() {
	p.locker.Lock()
	defer p.locker.Unlock()
	for p.poolCnt > 0 {
		conn, _ := p.Get()
		fmt.Println("关闭conn", conn)
		p.poolCnt -= 1
		p.count -= 1
	}
}

func main() {
	pool := New("127.0.0.1", 2, 5, 3)
	c1, _ := pool.Get()
	c2, _ := pool.Get()
	c3, _ := pool.Get()
	c4, _ := pool.Get()
	c5, _ := pool.Get()
	c6, err := pool.Get()

	fmt.Println(c1, c2, c3, c4, c5)
	fmt.Println(c6, err)

	pool.Put(c5)

	c6, err = pool.Get()
	fmt.Println(c6, err)
	pool.Put(c6)
	pool.Put(c4)
	pool.Put(c3)
	pool.Put(c2)
	pool.Put(c1)
}
