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

type (
	Connection struct {
		msgs   <-chan connectionMessage
		reqs   chan<- request
		events chan<- []byte

		m                sync.Mutex
		auth             authState
		err              error
		previousPacketID int32
		inflight         map[int32]*cmdHandle
		nextID           int32

		cancel    context.CancelFunc
		transport io.Closer
	}

	cmdHandle struct {
		body []byte
		// Either error or []byte
		res chan<- any
	}
)

const (
	minCmdID     int32 = 200
	authPacketID int32 = 11
)

func NewConnection(
	ctx context.Context,
	transport io.ReadWriteCloser,
	events chan<- []byte,
) *Connection {
	msgs := make(chan connectionMessage, 16)
	r := reader{
		src:  bufio.NewReader(transport),
		msgs: msgs,
	}

	reqs := make(chan request, 16)

	w := writer{
		reqs: reqs,
		dst:  transport,
		msgs: msgs,
	}

	ctx, cancel := context.WithCancel(ctx)

	conn := &Connection{
		msgs:      msgs,
		reqs:      reqs,
		events:    events,
		inflight:  make(map[int32]*cmdHandle, 16),
		nextID:    minCmdID,
		cancel:    cancel,
		transport: transport,
	}

	go r.nun(ctx)
	go w.run(ctx)
	go conn.Run(ctx)

	return conn
}

type connectionMessage interface {
	connectionMessage()
}

type authState int8

const (
	authStateInitial authState = iota
	authStateSentPacket
	authStateRecvResponseValue
	authStateDone
)

func (c *Connection) Auth(password string) error {
	ch, err := c.tryAuth(password)
	if err != nil {
		return err
	}

	res, ok := <-ch
	if !ok {
		return errors.New("connection was closed")
	}

	switch res := res.(type) {
	case error:
		return res
	case []byte:
		return nil
	default:
		panic(fmt.Errorf("unknown auth response type %T", res))
	}
}

func (c *Connection) IsHealthy() bool {
	c.m.Lock()
	ok := c.err == nil
	c.m.Unlock()
	return ok
}

func (c *Connection) tryAuth(password string) (chan any, error) {
	c.m.Lock()
	defer c.m.Unlock()
	if c.err != nil {
		return nil, c.err
	}

	switch c.auth {
	case authStateInitial:
		c.auth = authStateSentPacket
		authReq := newRequest(authPacketID, serverDataAuthTy, []byte(password))
		ch := make(chan any)
		c.inflight[authPacketID] = &cmdHandle{
			body: []byte{},
			res:  ch,
		}
		go func() {
			c.reqs <- authReq
		}()
		return ch, nil

	default:
		return nil, fmt.Errorf("authentication already started")
	}
}

func (c *Connection) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return

		case msg := <-c.msgs:
			if c.onMessage(msg) != nil {
				return
			}
		}
	}
}

func (c *Connection) onMessage(msg connectionMessage) error {
	c.m.Lock()
	defer c.m.Unlock()
	var err error

	switch msg := msg.(type) {
	case packet:
		err = c.onPacket(msg)
	case emptyPacket:
		err = c.onEmpty()
	case readerError:
		err = msg.Err
	case writerError:
		err = msg.Err
	default:
		panic(fmt.Errorf("unknown message type %T", msg))
	}
	if err == nil {
		return nil
	}

	return errors.Join(err, c.close(err))
}

func (c *Connection) onEmpty() error {
	h, exists := c.inflight[c.previousPacketID]
	if !exists {
		return fmt.Errorf("no inflight command with packetID %v", c.previousPacketID)
	}

	delete(c.inflight, c.previousPacketID)
	h.res <- h.body

	return nil
}

func (c *Connection) onPacket(p packet) error {

	switch c.auth {
	case authStateInitial:
		return fmt.Errorf("unexpected packet before auth id: %v\tty: %v", p.id, p.ty)
	case authStateSentPacket:
		if p.id == authPacketID && p.ty == serverDataResponseValueTy && len(p.body) == 0 {
			c.auth = authStateRecvResponseValue
			return nil
		}
		return fmt.Errorf("invalid auth SERVERDATA_RESPONSE_VALUE packet")

	case authStateRecvResponseValue:
		if p.id == authPacketID && p.ty == serverDataAuthResponseTy && len(p.body) == 0 {
			c.auth = authStateDone

			h := c.inflight[authPacketID]
			delete(c.inflight, authPacketID)

			h.res <- []byte{}
			close(h.res)

			return nil
		}
		return fmt.Errorf("invalid auth SERVERDATA_AUTH_RESPONSE packet")

	case authStateDone:
		if p.IsEvent() {
			c.events <- p.body
			return nil
		}

		if p.ty != serverDataResponseValueTy {
			return fmt.Errorf("unexpected packet type %v", p.ty)
		}

		h, exists := c.inflight[p.id]
		if !exists {
			return UnknownPacketError{PacketID: p.id}
		}

		c.previousPacketID = p.id
		h.body = append(h.body, p.body...)
		return nil
	}

	return nil
}

type UnknownPacketError struct{ PacketID int32 }

func (e UnknownPacketError) Error() string {
	return fmt.Sprintf("unknown packet with id %v", e.PacketID)
}

func (c *Connection) Close() error {
	c.m.Lock()
	defer c.m.Unlock()

	return c.close(errors.New("Close() called"))
}

// Caller is responsible for locking c.m
func (c *Connection) close(err error) error {
	if c.err != nil {
		return nil
	}

	if c.cancel != nil {
		c.cancel()
		c.cancel = nil
	}

	close(c.events)
	c.err = ConnectionClosedError{Err: err}

	for k, h := range c.inflight {
		h.res <- c.err
		close(h.res)
		delete(c.inflight, k)
	}

	err = c.transport.Close()
	c.transport = nopCloser{}

	return err
}

type nopCloser struct{}

func (nopCloser) Close() error { return nil }

type ConnectionClosedError struct {
	Err error
}

func (e ConnectionClosedError) Error() string {
	return fmt.Sprintf("connection closed: %v", e.Err)
}

func (e ConnectionClosedError) Unwrap() error {
	return e.Err
}
