// SPDX-FileCopyrightText: 2026 koitra <koitra@rhoti.com>
//
// SPDX-License-Identifier: SSPL-1.0

package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"src.rhoti.com/koitra/squadrcon/v2"
)

func main() {
	if err := run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	ctx := context.Background()
	host := os.Getenv("RCON_HOST")
	if host == "" {
		return errors.New("RCON_HOST not set")
	}

	password := os.Getenv("RCON_PASSWORD")
	if password == "" {
		return errors.New("RCON_PASSWORD not set")
	}

	cmd := os.Args[1:]
	if len(cmd) == 0 {
		return errors.New("no command provided")
	}

	p := squadrcon.NewPool(ctx, Provider{password: password})
	defer func() {
		_ = p.Close()
	}()

	conn, err := p.Get(ctx, host)
	if err != nil {
		return fmt.Errorf("failed to get connection: %w", err)
	}

	fullCmd := strings.Join(cmd, " ")

	res, err := conn.Command(ctx, fullCmd)
	if err != nil {
		return fmt.Errorf("failed to execute command: %w", err)
	}
	fmt.Println(res)

	return nil
}

type Provider struct {
	password string
}

func (p Provider) Get(ctx context.Context, host string) (squadrcon.ServerOpts, error) {
	sink := make(chan []byte, 64)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case _, ok := <-sink:
				if !ok {
					return
				}
			}
		}
	}()

	return squadrcon.ServerOpts{
		Passowrd:  p.password,
		EventSink: sink,
	}, nil
}
