// SPDX-FileCopyrightText: 2026 koitra <koitra@rhoti.com>
//
// SPDX-License-Identifier: SSPL-1.0

package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"time"

	"src.rhoti.com/koitra/squadrcon/v2"
)

func main() {
	if err := run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
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

	events := make(chan string, 32)
	go func() {
		for range events {
		}
	}()
	conn := squadrcon.Open(
		ctx,
		func(ctx context.Context) (io.ReadWriteCloser, error) {
			var d net.Dialer
			ctx, cancel := context.WithDeadline(ctx, time.Now().Add(time.Second*5))
			defer cancel()
			conn, err := d.DialContext(ctx, "tcp", host)
			if err != nil {
				return nil, fmt.Errorf("connect to host: %w", err)
			}
			return conn, nil
		},
		password,
		events,
	)

	fullCmd := strings.Join(cmd, " ")

	cmdCtx, cancel := context.WithDeadline(ctx, time.Now().Add(time.Second*10))
	defer cancel()
	res, err := conn.Command(cmdCtx, fullCmd)
	if err != nil {
		return fmt.Errorf("failed to execute command: %w", err)
	}
	fmt.Println(res)

	return nil
}
