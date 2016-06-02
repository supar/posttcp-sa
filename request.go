package main

import (
	"bytes"
	"errors"
	"net/url"
)

type Command struct {
	wr    bool
	value []byte
}

var CmdError = errors.New("read: requested unknown action")

func NewCommand(buf []byte) (cmd *Command, err error) {
	var (
		conv_str string
	)
	cmd = &Command{
		value: bytes.ToLower(buf),
	}

	if conv_str, err = url.QueryUnescape(string(cmd.value)); err != nil {
		return nil, err
	} else {
		cmd.value = []byte(conv_str)
	}

	if err = cmd.init(); err != nil {
		return nil, err
	}

	cmd.clear()

	return
}

func (this *Command) GetStr() string {
	return string(this.value)
}

func (this *Command) init() error {
	var (
		prefix = [][]byte{
			[]byte("get "),
			[]byte("put "),
		}
	)

	for idx, val := range prefix {
		if bytes.HasPrefix(this.value, val) {
			this.value = bytes.TrimPrefix(this.value, val)

			switch idx {
			case 1:
				this.wr = true
			default:
				this.wr = false
			}

			return nil
		}
	}

	return CmdError
}

func (this *Command) clear() {
	for {
		if b, ok := hasGarbage(this.value); ok {
			this.value = bytes.TrimPrefix(this.value, b)
			this.value = bytes.TrimSuffix(this.value, b)

			continue
		}

		break
	}
}

func hasGarbage(b []byte) ([]byte, bool) {
	var (
		garbage = []byte{
			0x00,
			0x0a,
			0x0d,
			0x20,
		}
	)

	for _, g := range garbage {
		if bytes.HasPrefix(b, []byte{g}) || bytes.HasSuffix(b, []byte{g}) {
			return []byte{g}, true
		}
	}

	return nil, false
}
