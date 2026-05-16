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
		active   map[string]*Connection
		provider Provider
		ctx      context.Context
	}

	Provider interface {
		Get(context.Context, string) (ServerOpts, error)
	}

	ServerOpts struct {
		Passowrd  string
		EventSink chan<- []byte
	}
)

func NewPool(
	ctx context.Context,
	provider Provider,
) *ServerPool {
	return &ServerPool{
		mu:       sync.Mutex{},
		active:   map[string]*Connection{},
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

	conn, ok := p.active[host]
	if ok && conn.IsHealthy() {
		return conn, nil
	}

	tcpCon, err := net.Dial("tcp", host)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to server: %w", err)
	}

	opts, err := p.provider.Get(p.ctx, host)
	if err != nil {
		return nil, fmt.Errorf("failed to get server opts: %w", err)
	}

	conn = NewConnection(
		p.ctx,
		tcpCon,
		opts.EventSink,
	)

	err = conn.Auth(opts.Passowrd)
	if err != nil {
		return nil, fmt.Errorf("failed to setup connection: %w", err)
	}

	p.active[host] = conn

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

	conn, exists := p.active[host]
	if exists {
		_ = conn.Close()
		delete(p.active, host)
	}
}

func (p *ServerPool) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	var closeErr []error
	for _, conn := range p.active {
		err := conn.Close()
		if closeErr != nil {
			closeErr = append(closeErr, err)
		}
	}

	return errors.Join(closeErr...)
}
