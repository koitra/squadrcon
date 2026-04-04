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

type EventSinkFactory = func(context.Context, Hostname) (chan<- []byte, error)
type Hostname = string

type Pool struct {
	mu            sync.Mutex
	active        map[string]*Connection
	makeEventSink EventSinkFactory
	ctx           context.Context
}

func NewPool(
	ctx context.Context,
	makeEventSink EventSinkFactory,
) *Pool {
	return &Pool{
		mu:            sync.Mutex{},
		active:        map[string]*Connection{},
		makeEventSink: makeEventSink,
		ctx:           ctx,
	}
}

func (p *Pool) Get(
	_ context.Context, host string, password string,
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

	events, err := p.makeEventSink(p.ctx, host)
	if err != nil {
		return nil, fmt.Errorf("failed to create event sink: %w", err)
	}

	conn, err = NewConnection(
		p.ctx,
		tcpCon,
		password,
		events,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection: %w", err)
	}

	p.active[host] = conn

	return conn, nil
}

func (p *Pool) Reconnect(
	ctx context.Context, host string, password string,
) (*Connection, error) {
	p.Stop(ctx, host)
	return p.Get(ctx, host, password)
}

func (p *Pool) Stop(ctx context.Context, host string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	conn, exists := p.active[host]
	if exists {
		_ = conn.Close()
		delete(p.active, host)
	}
}

func (p *Pool) Close() error {
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
