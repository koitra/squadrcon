// SPDX-FileCopyrightText: 2026 koitra <koitra@rhoti.com>
//
// SPDX-License-Identifier: SSPL-1.0

package events

import (
	"fmt"

	"github.com/oriser/regroup"
)

type KickedPlayer struct {
	ConnectID int32
	EosID     string
	SteamID   string
	Name      string
}

var kickedPlayerRe = regroup.MustCompile(
	`Kicked player (?P<connectID>\d+)\. \[Online IDs= EOS: (?P<eosID>[\da-z]{32}) steam: (?P<steamID>\d+)\]\s+(?P<name>.*)$`,
)

func (e *KickedPlayer) parse(b []byte) error {
	type record struct {
		ConnectID int32  `regroup:"connectID"`
		EosID     string `regroup:"eosID"`
		SteamID   string `regroup:"steamID"`
		Name      string `regroup:"name"`
	}

	var r record
	err := kickedPlayerRe.MatchToTarget(string(b), &r)
	if err != nil {
		return fmt.Errorf("match failed: %w", err)
	}

	e.ConnectID = r.ConnectID
	e.EosID = r.EosID
	e.SteamID = r.SteamID
	e.Name = r.Name

	return nil
}
