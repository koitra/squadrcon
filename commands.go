// SPDX-FileCopyrightText: 2026 koitra <koitra@rhoti.com>
//
// SPDX-License-Identifier: SSPL-1.0

package squadrcon

import (
	"context"

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

func (c *Connection) ShowServerInfo(ctx context.Context) (commands.ShowServerInfoResponse, error) {
	return runCmd(ctx, c, commands.ShowServerInfo{})
}

func runCmd[Command commands.RconCommand[Response], Response any](ctx context.Context, c *Connection, cmd Command) (Response, error) {
	body, err := c.Command(ctx, cmd.ToBody())
	if err != nil {
		return *new(Response), err
	}

	return cmd.ParseBody(body)
}
