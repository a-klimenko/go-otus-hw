package main

import (
	"errors"
	"fmt"
	"io"
	"net"
	"time"
)

const ConnType = "tcp"

var (
	ErrInvalidIn  = errors.New("in is invalid")
	ErrInvalidOut = errors.New("out is invalid")
)

type TelnetClient interface {
	Connect() error
	io.Closer
	Send() error
	Receive() error
}

func NewTelnetClient(address string, timeout time.Duration, in io.ReadCloser, out io.Writer) TelnetClient {
	return &Client{
		address: address,
		timeout: timeout,
		in:      in,
		out:     out,
	}
}

type Client struct {
	address string
	timeout time.Duration
	in      io.ReadCloser
	out     io.Writer
	conn    net.Conn
}

func (c *Client) Connect() error {
	var err error

	if c.in == nil {
		return ErrInvalidIn
	}
	if c.out == nil {
		return ErrInvalidOut
	}

	c.conn, err = net.DialTimeout(ConnType, c.address, c.timeout)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) Close() error {
	if err := c.conn.Close(); err != nil {
		return err
	}

	return nil
}

func (c *Client) Send() error {
	if _, err := io.Copy(c.conn, c.in); err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}

func (c *Client) Receive() error {
	if _, err := io.Copy(c.out, c.conn); err != nil {
		return fmt.Errorf("failed to receive message: %w", err)
	}

	return nil
}
