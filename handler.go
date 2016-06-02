package main

import (
	"bufio"
	"bytes"
	"net"
)

type Connection struct {
	id string
	*net.TCPConn
}

func NewConnection(c *net.TCPConn) *Connection {
	return &Connection{
		id:      RandStringId(12),
		TCPConn: c,
	}
}

func (this *Connection) Id() string {
	return this.id
}

func (this *Connection) SendError() (err error) {
	_, err = this.send("400", "")
	return
}

func (this *Connection) SendNotFound() (err error) {
	_, err = this.send("500", "")
	return
}

func (this *Connection) SendReject() (err error) {
	_, err = this.send("200", "REJECT")
	return
}

func (this *Connection) send(code, msg string) (int, error) {
	var buf = bytes.NewBufferString(code + " " + msg + "\n")

	return this.Write(buf.Bytes())
}

func readConn(conn *Connection, callback_fn func(buf []byte) (int64, error)) (err error) {
	var (
		buf     []byte
		count   int64
		scanner *bufio.Scanner
	)

	scanner = bufio.NewScanner(conn)

	// Read from connection with line split
	for scanner.Scan() {
		buf = scanner.Bytes()

		count, err = callback_fn(buf)

		if err != nil {
			if err_s := conn.SendError(); err_s != nil {
				return err_s
			}

			continue
		}

		if count == 0 {
			if err = conn.SendNotFound(); err != nil {
				return
			}
		} else {
			if err = conn.SendReject(); err != nil {
				return
			}
		}
	}

	return scanner.Err()
}
