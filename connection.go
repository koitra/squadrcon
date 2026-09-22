// SPDX-FileCopyrightText: 2026 koitra <koitra@rhoti.com>
//
// SPDX-License-Identifier: SSPL-1.0

package squadrcon

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
)

type Connection struct {
	dial     Dialer
	events   chan<- string
	passowrd string

	mu      sync.Mutex
	pending []command
	conn    *connection
}

type Dialer = func(ctx context.Context) (io.ReadWriteCloser, error)

func Open(ctx context.Context, dial Dialer, password string, events chan<- string) *Connection {
	c := &Connection{
		dial:     dial,
		events:   events,
		passowrd: password,
	}
	go c.run(ctx)
	return c
}

func (c *Connection) run(ctx context.Context) {
	for {
		var (
			pending []command
			err     error
		)

		bus := make(chan any, 32)
		withLock(&c.mu, func() {
			var t io.ReadWriteCloser
			t, err = c.dial(ctx)
			if err != nil {
				for _, p := range c.pending {
					select {
					case p.result <- errConnectionClosed:
					default:
					}
				}
				c.pending = []command{}
				return
			}

			conn := connection{
				t:          t,
				bus:        bus,
				nextID:     1000,
				events:     c.events,
				inflight:   make(map[int32]*cmdHandle),
				previousID: 0,
			}
			c.conn = &conn
			pending = c.pending
			c.pending = []command{}
		})
		go func() {
			for _, p := range pending {
				if p.ctx.Err() != nil {
					continue
				}
				bus <- p
			}
		}()
		if err != nil {
			continue
		}
		err = c.conn.run(ctx, c.passowrd)
		if errors.Is(err, context.Canceled) {
			return
		}
	}
}

func (c *Connection) Command(ctx context.Context, cmd string) (string, error) {
	var err error
	n := command{
		body:   cmd,
		ctx:    ctx,
		result: make(chan any),
	}
	var errNoActiveConnection = errors.New("no active connection")

	withLock(&c.mu, func() {
		if c.conn == nil {
			c.pending = append(c.pending, n)
			err = errNoActiveConnection
			return
		}
		select {
		case <-ctx.Done():
			err = ctx.Err()
		case c.conn.bus <- n:
		}
	})
	if err != nil && !errors.Is(err, errNoActiveConnection) {
		return "", err
	}

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case res := <-n.result:
		switch res := res.(type) {
		case error:
			return "", res
		case []byte:
			return string(res), nil
		default:
			return "", fmt.Errorf("unknown result: %+v", res)
		}
	}
}

type (
	connection struct {
		t   io.ReadWriteCloser
		bus chan any // command / packet / emptyPacket / readError

		nextID     int32
		events     chan<- string
		inflight   map[int32]*cmdHandle
		previousID int32
	}

	cmdHandle struct {
		body   []byte
		ctx    context.Context
		result chan any // []byte or error
	}

	command struct {
		body   string
		ctx    context.Context
		result chan any // []byte or error
	}
)

func (c *connection) run(ctx context.Context, password string) error {
	defer func() {
		for h := range c.bus {
			cmd, ok := h.(command)
			if !ok {
				continue
			}
			select {
			case cmd.result <- errConnectionClosed:
			default:
			}
		}
	}()

	if err := c.authenticate(password); err != nil {
		return err
	}

	r := reader{
		src: bufio.NewReader(c.t),
		bus: c.bus,
	}

	readerCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	go r.run(readerCtx)

	for {
		select {
		case <-ctx.Done():
			err := ctx.Err()
			defer c.cleanup(err)
			return err
		case m := <-c.bus:
			err := c.onMessage(m)
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				continue
			}
			if err != nil {
				defer c.cleanup(err)
				return err
			}
		}
	}
}

func (c *connection) authenticate(password string) error {
	r := newRequest(authPacketID, serverDataAuthTy, []byte(password))
	err := r.write(c.t)
	if err != nil {
		return writeError{err}
	}

	p, err := readPacket(c.t)
	if err != nil {
		return fmt.Errorf("SERVERDATA_AUTH_RESPONSE_VALUE packet: %w", readError{err})
	}
	if !p.IsAuthResponseValue() {
		return readError{
			fmt.Errorf("invalid SERVERDATA_AUTH_RESPONSE_VALUE packet: %+v", p),
		}
	}

	p, err = readPacket(c.t)
	if err != nil {
		return fmt.Errorf("SERVERDATA_AUTH_RESPONSE packet: %w", readError{err})
	}
	if !p.IsAuthResponse() {
		return readError{
			fmt.Errorf("invalid SERVERDATA_AUTH_RESPONSE: %+v", p),
		}
	}

	return nil
}

func (c *connection) onMessage(m any) error {
	switch m := m.(type) {
	case error:
		return m
	case emptyPacket:
		return c.onEmpty()
	case packet:
		return c.onPacket(m)
	case command:
		return c.onCommand(m)
	default:
		return fmt.Errorf("unknown item in bus: %+v", m)
	}
}

func (c *connection) onEmpty() error {
	h, ok := c.inflight[c.previousID]
	if !ok {
		return fmt.Errorf("unknown command with id %v", c.previousID)
	}
	delete(c.inflight, c.previousID)
	defer close(h.result)
	if h.ctx.Err() != nil {
		return h.ctx.Err()
	}
	select {
	case h.result <- h.body:
	default:
	}
	return nil
}

func (c *connection) onPacket(p packet) error {
	if p.IsEvent() {
		select {
		case c.events <- string(p.body):
		default:
		}
		return nil
	}

	h, ok := c.inflight[p.id]
	if !ok {
		return fmt.Errorf("unknown command with id %v", p.id)
	}
	h.body = append(h.body, p.body...)
	c.previousID = p.id
	return nil
}

func (c *connection) onCommand(cmd command) error {
	if cmd.ctx.Err() != nil {
		return cmd.ctx.Err()
	}

	id := c.nextID
	c.nextID += 1
	r := newRequest(id, serverDataExecCommandTy, []byte(cmd.body))
	r.appendEmpty()

	c.inflight[id] = &cmdHandle{
		result: cmd.result,
		ctx:    cmd.ctx,
	}

	if err := r.write(c.t); err != nil {
		return writeError{err}
	}

	return nil
}

func (c *connection) cleanup(err error) {
	for _, h := range c.inflight {
		select {
		case h.result <- ConnectionClosedError{err}:
		default:
		}
		close(h.result)
	}
	_ = c.t.Close()
}

type writeError struct {
	Err error
}

func (e writeError) Error() string {
	return fmt.Sprintf("write: %v", e.Err)
}

func (e writeError) Unwrap() error {
	return e.Err
}

type ConnectionClosedError struct {
	inner error
}

func (e ConnectionClosedError) Error() string {
	return fmt.Sprintf("%v", e.inner)
}

func (e ConnectionClosedError) Unwrap() error {
	return e.inner
}

var errConnectionClosed = errors.New("connection was closed")

func withLock(mu *sync.Mutex, f func()) {
	mu.Lock()
	defer mu.Unlock()
	f()
}
