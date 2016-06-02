package main

import (
	"bytes"
	"net"
	"testing"
)

type ReadConnMock struct {
	Request  []byte
	Found    int64
	Response []byte
}

func newLocalTcpListener() (*Server, error) {
	return NewServer("127.0.0.1:2222")
}

func newTCPDialer(l *Server) (*net.TCPConn, error) {
	local, _ := net.ResolveTCPAddr("tcp", "127.0.0.1")
	remote, _ := net.ResolveTCPAddr("tcp", l.Addr().String())
	return net.DialTCP("tcp", local, remote)
}

func Test_Command(t *testing.T) {
	var (
		mock = map[string][]byte{
			"test_data_a": []byte("get Test_data_A\n"),
			"test_data_b": []byte("get Test_data_B\r\n"),
			"test_data_c": []byte("get Test_data_C\r\n\n\n\r\n\n"),
			"test_data d": []byte("get Test_data%20D\n"),
			"test_data f": []byte("put Test_data%20F%20%20\r%20\n"),
		}
		req *Command
		err error
	)

	for i, val := range mock {
		if req, err = NewCommand(val); err != nil {
			t.Error("Unexpected error %s", err.Error())
		}

		if v := string(req.value); v != i {
			t.Errorf("Expected value %v, but got %v", []byte(i), req.value)
		}
	}
}

func Test_ReadConnectionWrap(t *testing.T) {
	var (
		mock = []ReadConnMock{
			ReadConnMock{
				Request:  []byte("get any.data_A\r\r\r\r\n"),
				Found:    0,
				Response: []byte("500 \n"),
			},
			ReadConnMock{
				Request:  []byte("iput any.data_P\n"),
				Found:    12,
				Response: []byte("200 OK \n"),
			},
		}

		mock_req = make([][]byte, 0)

		l   *Server
		cl  *net.TCPConn
		err error

		// Create sync channel
		done = make(chan bool)
	)

	if l, err = newLocalTcpListener(); err != nil {
		t.Fatal(err)
	}

	go func() {
		defer func() {
			l.Close()
			done <- true
		}()

		c, err := l.AcceptTCP()

		if err != nil {
			t.Fatal(err)
		}

		conn := NewConnection(c)
		err = readConn(conn, func(buf []byte) (int64, error) {
			mock_req = append(mock_req, buf)

			for _, v := range mock {
				if bytes.HasPrefix(v.Request, buf) {
					return v.Found, nil
				}
			}

			return 0, nil
		})

		if err != nil {

			if neterr, ok := err.(net.Error); !ok || !neterr.Temporary() {
				t.Error(err)
			}
		}
	}()

	go func() {
		if cl, err = newTCPDialer(l); err != nil {
			t.Fatal(err)
		}
		defer cl.Close()

		for _, v := range mock {
			if _, err = cl.Write(v.Request); err != nil {
				t.Error(err)
			}
		}
	}()

	<-done

	if len(mock) != len(mock_req) {
		t.Fatalf("Expected equal size of the write and read slice, but got w:%d<->r:%d", len(mock), len(mock_req))
	}

	for idx, v := range mock {
		if len(mock_req[idx]) == 0 || bytes.HasPrefix(v.Request, mock_req[idx]) == false {
			t.Errorf("Expected %v, but got %v", v, mock_req[idx])
		}
	}
}
