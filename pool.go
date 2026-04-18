// SPDX-FileCopyrightText: 2026 koitra <koitra@rhoti.com>
//
// SPDX-License-Identifier: SSPL-1.0

package squadrcon

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
)

type (
	Pool interface {
		Get(context.Context, string) (*Connection, error)
		Reconnect(context.Context, string) (*Connection, error)
		Stop(context.Context, string)
		Close() error
	}

	ServerPool struct {
		mu       sync.Mutex
		active   map[string]*handle
		provider Provider
		ctx      context.Context
	}

	Provider interface {
		Get(context.Context, string) (ServerOpts, error)
	}

	ServerOpts struct {
		Passowrd  string
		EventSink chan<- []byte
		OnClose   func()
	}

	handle struct {
		conn    *Connection
		onClose func()
	}
)

func NewPool(
	ctx context.Context,
	provider Provider,
) *ServerPool {
	return &ServerPool{
		mu:       sync.Mutex{},
		active:   map[string]*handle{},
		provider: provider,
		ctx:      ctx,
	}
}

func (p *ServerPool) Get(
	_ context.Context,
	host string,
) (*Connection, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	existing, ok := p.active[host]
	conn := existing.conn
	if ok && conn.IsHealthy() {
		return conn, nil
	}
	go existing.onClose()
	h := new(handle)
	var err error
	defer func() {
		if err != nil {
			_ = h.Close()
		}
	}()

	tcpCon, err := net.Dial("tcp", host)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to server: %w", err)
	}

	opts, err := p.provider.Get(p.ctx, host)
	if err != nil {
		return nil, fmt.Errorf("failed to get server opts: %w", err)
	}
	h.onClose = opts.OnClose
	if h.onClose == nil {
		h.onClose = func() {}
	}

	conn, err = NewConnection(
		p.ctx,
		tcpCon,
		opts.Passowrd,
		opts.EventSink,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create connection: %w", err)
	}

	h.conn = conn
	p.active[host] = h

	return conn, nil
}

func (p *ServerPool) Reconnect(
	ctx context.Context,
	host string,
) (*Connection, error) {
	p.Stop(ctx, host)
	return p.Get(ctx, host)
}

func (p *ServerPool) Stop(ctx context.Context, host string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	h, exists := p.active[host]
	if exists {
		_ = h.conn.Close()
		delete(p.active, host)
		go h.onClose()
	}
}

func (p *ServerPool) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	var closeErr []error
	for _, h := range p.active {
		err := h.Close()
		if closeErr != nil {
			closeErr = append(closeErr, err)
		}
		go h.onClose()
	}

	return errors.Join(closeErr...)
}

func (h *handle) Close() error {
	var err error
	if h.conn != nil {
		err = h.conn.Close()
		h.conn = nil
	}

	if h.onClose != nil {
		go h.onClose()
	}

	return err
}
