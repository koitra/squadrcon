// SPDX-FileCopyrightText: 2026 koitra <koitra@rhoti.com>
//
// SPDX-License-Identifier: SSPL-1.0

package events

import (
	"fmt"
	"strconv"

	"github.com/oriser/regroup"
)

type CreatedSquad struct {
	PlayerName string

	EosID   string
	SteamID string

	SquadName   string
	SquadNumber int

	Team string
}

var createdSquadRe = regroup.MustCompile(
	`(?P<player>.+) \(Online IDs: EOS: (?P<eosID>.*) steam: (?P<steamID>\d+)\) has created Squad (?P<squadNumber>\d+) \(Squad Name: (?P<squadName>.+)\) on (?P<team>.+)$`,
)

func (e *CreatedSquad) parse(b []byte) error {
	type record struct {
		PlayerName  string `regroup:"player"`
		EosID       string `regroup:"eosID"`
		SteamID     string `regroup:"steamID"`
		SquadNumber string `regroup:"squadNumber"`
		SquadName   string `regroup:"squadName"`
		Team        string `regroup:"team"`
	}

	var r record
	err := createdSquadRe.MatchToTarget(string(b), &r)
	if err != nil {
		return fmt.Errorf("match failed: %w", err)
	}

	e.PlayerName = r.PlayerName
	e.EosID = r.EosID
	e.SteamID = r.SteamID
	e.SquadName = r.SquadName
	e.SquadNumber, err = strconv.Atoi(r.SquadNumber)
	if err != nil {
		// \d is used in the regex, so it should be a number
		panic(fmt.Errorf("squad number is not a number: %v", err))
	}
	e.Team = r.Team

	return nil
}
