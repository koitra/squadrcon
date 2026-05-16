// SPDX-FileCopyrightText: 2026 koitra <koitra@rhoti.com>
//
// SPDX-License-Identifier: SSPL-1.0

package squadrcon

import (
	"context"
	"errors"
	"fmt"
	"unicode/utf8"

	"src.rhoti.com/koitra/squadrcon/v2/commands"
)

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

func (c *Connection) Command(ctx context.Context, command string) (string, error) {
	ch, err := c.cmd(command)
	if err != nil {
		return "", err
	}

	res, ok := <-ch

	if !ok {
		return "", errors.New("connection closed")
	}

	switch res := res.(type) {
	case error:
		return "", res
	case []byte:
		valid := utf8.Valid(res)
		if !valid {
			return "", errors.New("response string is not valid utf8")
		}

		return string(res), nil

	default:
		panic(fmt.Errorf("unknown message type %T", res))
	}
}

func (c *Connection) ShowServerInfo(ctx context.Context) (commands.ShowServerInfoResponse, error) {
	return runCmd(ctx, c, commands.ShowServerInfo{})
}

func (c *Connection) cmd(command string) (chan any, error) {
	c.m.Lock()
	defer c.m.Unlock()

	if c.err != nil {
		return nil, c.err
	}

	if c.auth != authStateDone {
		return nil, errors.New("not authenticated")
	}

	id := c.nextID
	c.nextID += 1

	ch := make(chan any)
	c.inflight[id] = &cmdHandle{
		body: []byte{},
		res:  ch,
	}

	req := newRequest(id, serverDataExecCommandTy, []byte(command))
	req.appendEmpty()
	go func() {
		c.reqs <- req
	}()

	return ch, nil
}

func runCmd[Command commands.RconCommand[Response], Response any](ctx context.Context, c *Connection, cmd Command) (Response, error) {
	body, err := c.Command(ctx, cmd.ToBody())
	if err != nil {
		return *new(Response), err
	}

	return cmd.ParseBody(body)
}
