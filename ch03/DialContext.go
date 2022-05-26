package ch03

import (
	"context"
	"net"
	"sync"
	"testing"
	"time"
)

func TestDialContextCancelFanOut(t *testing.T) {
	ctx, cancel := context.WithDeadline(
		context.Background(),
		time.Now().Add(10*time.Second),
	)
	listener, err := net.Listen("tcp", "127.0.0.1:")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	go func() {
		//하나의 연결만을 수락
		conn, err := listener.Accept()
		if err == nil {
			conn.Close()
		}
	}()

	dial := func(ctx context, address string, response chan int,
		id int, wg *sync.WaitGroup) {
		defer wg.Done()
		var d net.Dialer
		c, err := d.DialContext(ctx, "tcp", address)
		if err != nil {
			return
		}

		c.Close()
	}

	select {
	case <-ctx.Done():
	case response <- id:
	}
	res := make(chan int)
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(i)
		go dial(ctx, listener.Addr().string, res, i+1, &wg)
	}

	response := <-res
	cancel()
	wg.Wait()
	close()

	if ctx.Err() != context.Canceled {
		t.Errorf("expected cancled context; actual:%s", ctx.Err())
	}
	t.Logf("dialer %d retrieved the resource", response)
}
