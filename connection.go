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
	"sync/atomic"
	"unicode/utf8"

	"src.rhoti.com/koitra/squadrcon/commands"
)

type (
	Connection struct {
		transport io.Closer

		rh readerHandle
		wh writerHandle

		events chan<- []byte

		inflight         map[int32]*cmdHandle
		nextID           int32
		previousPacketID int32

		newCmds chan newCommand

		ctx context.Context

		isDead atomic.Bool
		error  error

		cancel context.CancelFunc
	}

	newCommand struct {
		body string
		res  chan<- []byte
		err  chan<- error
	}

	cmdHandle struct {
		body []byte
		res  chan<- []byte
		err  chan<- error
	}

	readerHandle struct {
		packets <-chan packet
		empty   <-chan struct{}
		err     <-chan error
	}

	writerHandle struct {
		reqs chan<- request
		err  <-chan error
	}

	UnknownPacketError struct{ PacketID int32 }
)

const (
	minCmdID     int32 = 200
	authPacketID int32 = 11
)

func (c *Connection) run() {
	// TODO: ping
	for {
		select {
		case <-c.ctx.Done():
			c.degrade(c.ctx.Err())
			return

		case err := <-c.rh.err:
			c.degrade(fmt.Errorf("reader error: %w", err))
			return

		case packet := <-c.rh.packets:
			err := c.onPacket(packet)
			if err != nil {
				c.degrade(fmt.Errorf("failed to handle packet: %w", err))
				return
			}

		case <-c.rh.empty:
			err := c.onEmpty()
			if err != nil {
				c.degrade(fmt.Errorf("failed to handle empty packet: %w", err))
				return
			}

		case err := <-c.wh.err:
			c.degrade(fmt.Errorf("writer error: %w", err))
			return

		case cmd := <-c.newCmds:
			c.onNewCommand(cmd)
		}
	}
}

func (c *Connection) onNewCommand(cmd newCommand) {
	id := c.nextID
	c.inflight[id] = &cmdHandle{
		body: []byte{},
		res:  cmd.res,
		err:  cmd.err,
	}

	// TODO: handle overflow
	c.nextID += 1

	req := newRequest(id, serverDataExecCommandTy, []byte(cmd.body))
	req.packets = append(req.packets, packet{
		id:   id,
		ty:   serverDataExecCommandTy,
		body: []byte{},
	})

	c.wh.reqs <- req
}

func (c *Connection) onPacket(p packet) error {
	if p.IsEvent() {
		c.events <- p.body
		return nil
	}

	if p.ty != serverDataResponseValueTy {
		return UnknownPacketError{PacketID: p.id}
	}

	h, exists := c.inflight[p.id]
	if !exists {
		return UnknownPacketError{PacketID: p.id}
	}

	c.previousPacketID = p.id
	h.body = append(h.body, p.body...)
	return nil
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

func startReader(src io.Reader, ctx context.Context) readerHandle {
	packets := make(chan packet, 16)
	empty := make(chan struct{}, 16)
	error := make(chan error, 1)
	hdl := readerHandle{
		packets: packets,
		empty:   empty,
		err:     error,
	}

	buf := bufio.NewReader(src)

	go func() {
		for ctx.Err() == nil {
			end, err := readEmptyPacket(buf)
			if err != nil && !errors.Is(err, notAnEmptyPacketError{}) {
				error <- fmt.Errorf("failed to read empty packet: %w", err)
				return
			}
			if err == nil {
				empty <- end
				continue
			}

			packet, err := readPacket(buf)
			if err != nil {
				error <- fmt.Errorf("failed to read packet: %w", err)
				return
			}

			packets <- packet
		}
	}()

	return hdl
}

func startWriter(dst io.Writer, ctx context.Context) writerHandle {
	requests := make(chan request, 8)
	error := make(chan error)
	hdl := writerHandle{
		reqs: requests,
		err:  error,
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return

			case req := <-requests:
				err := req.write(dst)
				if err != nil {
					error <- fmt.Errorf("failed to write request: %w", err)
					return
				}
			}
		}
	}()

	return hdl
}

func NewConnection(
	ctx context.Context,
	transport io.ReadWriteCloser,
	password string,
	events chan<- []byte,
) (*Connection, error) {
	authReq := newRequest(authPacketID, serverDataAuthTy, []byte(password))
	err := authReq.write(transport)
	if err != nil {
		return nil, fmt.Errorf("failed to write auth packet: %w", err)
	}

	_, err = readPacket(transport)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to read empty response value packet: %w",
			err,
		)
	}

	_, err = readPacket(transport)
	if err != nil {
		return nil, fmt.Errorf("failed to read auth resp packet: %w", err)
	}

	rh := startReader(transport, ctx)
	wh := startWriter(transport, ctx)

	ctx, cancel := context.WithCancel(ctx)

	con := Connection{
		transport:        transport,
		rh:               rh,
		wh:               wh,
		events:           events,
		inflight:         make(map[int32]*cmdHandle),
		nextID:           minCmdID,
		previousPacketID: -1,
		newCmds:          make(chan newCommand),
		ctx:              ctx,
		isDead:           atomic.Bool{},
		error:            nil,
		cancel:           cancel,
	}

	go con.run()

	return &con, nil
}

func (c *Connection) degrade(err error) {
	if !c.isDead.CompareAndSwap(false, true) {
		return
	}

	c.error = err
	for _, hdl := range c.inflight {
		hdl.err <- err
	}

	_ = c.transport.Close()
	close(c.events)
}

func (c *Connection) Command(ctx context.Context, body string) (string, error) {
	res := make(chan []byte, 1)
	err := make(chan error, 1)

	c.newCmds <- newCommand{
		body: body,
		res:  res,
		err:  err,
	}

	select {
	case <-ctx.Done():
		return "", ctx.Err()

	case body := <-res:
		if utf8.Valid(body) {
			return string(body), nil
		}
		return "", fmt.Errorf("invalid utf8 in response `%v`", body)

	case err := <-err:
		return "", err
	}
}

func (c *Connection) Close() error {
	c.cancel()

	return nil
}

func (e UnknownPacketError) Error() string {
	return fmt.Sprintf("unknown packet with id %v", e.PacketID)
}

func (c *Connection) Status() ConnectionStatus {
	if c.isDead.Load() {
		return StatusDegraded
	}

	return StatusHealthy
}

func (c *Connection) IsHealthy() bool {
	return c.Status() == StatusHealthy
}

func (c *Connection) ListPlayers(ctx context.Context) (commands.ListPlayersResponse, error) {
	return runCmd(ctx, c, commands.ListPlayers{})
}

func (c *Connection) AdminKick(
	ctx context.Context,
	player string,
	reason string,
) (commands.AdminKickResponse, error) {
	return runCmd(ctx, c, commands.AdminKick{
		Player: player,
		Reason: reason,
	})
}

func (c *Connection) ListSquads(ctx context.Context) (commands.ListSquadsResponse, error) {
	return runCmd(ctx, c, commands.ListSquads{})
}

func (c *Connection) AdminWarnByID(
	ctx context.Context,
	eosID string,
	message string,
) (commands.AdminWarnByIDResponse, error) {
	return runCmd(ctx, c, commands.AdminWarnByID{
		EosID:   eosID,
		Message: message,
	})
}

func runCmd[Command commands.RconCommand[Response], Response any](
	ctx context.Context,
	conn *Connection,
	cmd Command,
) (Response, error) {
	body, err := conn.Command(ctx, cmd.ToBody())
	if err != nil {
		return *new(Response), fmt.Errorf("command error: %w", err)
	}

	res, err := cmd.ParseBody(body)
	if err != nil {
		return *new(Response), err
	}

	return res, nil
}
