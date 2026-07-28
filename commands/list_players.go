// SPDX-FileCopyrightText: 2026 koitra <koitra@rhoti.com>
//
// SPDX-License-Identifier: SSPL-1.0

package commands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/oriser/regroup"
)

type ListPlayers struct{}

func (ListPlayers) ToBody() string { return "ListPlayers" }
func (ListPlayers) ParseBody(body string) (ListPlayersResponse, error) {
	return parseListPlayers(body)
}

type (
	ListPlayersResponse struct {
		Active       []*ActivePlayer
		Disconnected []*DisconnectedPlayer
	}

	ActivePlayer struct {
		ConnectID   int32
		EosID       string
		SteamID     *string
		Name        string
		TeamNumber  *int32
		SquadNumber *int32
		IsLeader    bool
		Role        string
	}

	activePlayerRecord struct {
		ID        int32  `regroup:"connect_id"`
		EosID     string `regroup:"eos_id"`
		SteamID   string `regroup:"steam_id"`
		WithSteam bool   `regroup:"steam_id,exists"`
		Name      string `regroup:"name"`
		TeamID    string `regroup:"team_id"`
		SquadID   string `regroup:"squad_id"`
		IsLeader  bool   `regroup:"is_leader"`
		Role      string `regroup:"role"`
	}

	DisconnectedPlayer struct {
		ConnectID        int32
		EosID            string
		SteamID          *string
		Name             string
		DisconnectedSecs int32
	}

	disconnectedPlayerRecord struct {
		ConnectID int32  `regroup:"connect_id"`
		EosID     string `regroup:"eos_id"`
		WithSteam bool   `regroup:"steam_id,exists"`
		SteamID   string `regroup:"steam_id"`
		Name      string `regroup:"name"`
		Sec       int32  `regroup:"disconnect_sec"`
		Min       int32  `regroup:"disconnect_min"`
	}
)

var (
	activePlayerRe = regroup.MustCompile(
		`^ID: (?P<connect_id>\d+) \| Online IDs: EOS: (?P<eos_id>[\da-z]{32})( steam: (?P<steam_id>\d+))? \| Name:\s+(?P<name>.*) \| Team ID: (?P<team_id>(\d)|(N/A)) \| Squad ID: (?P<squad_id>(\d+)|(N/A)) \| Is Leader: (?P<is_leader>(False)|(True)) \| Role: (?P<role>.*)$`,
	)

	disconnectedPlayerRe = regroup.MustCompile(
		`ID: (?P<connect_id>\d+) \| Online IDs: EOS: (?P<eos_id>[\da-z]{32})( steam: (?P<steam_id>\d+))? \| Since Disconnect: (?P<disconnect_min>\d+)m\.(?P<disconnect_sec>\d+)s \| Name:\s+(?P<name>.*)$`,
	)
)

const (
	activeHeader       = "----- Active Players -----"
	disconnectedHeader = "----- Recently Disconnected Players [Max of 15] -----"
)

func parseListPlayers(body string) (ListPlayersResponse, error) {
	normalized := strings.ReplaceAll(body, "\r\n", "\n")
	normalized = strings.ReplaceAll(normalized, "\r", "\n")
	lines := strings.Lines(normalized)
	type state = int8
	const (
		ready state = iota
		inActive
		inDisconnected
	)

	s := ready

	var activePlayers []*ActivePlayer
	var disconnectedPlayers []*DisconnectedPlayer

	for line := range lines {
		line = strings.TrimSpace(line)
		switch s {
		case ready:
			if line != activeHeader {
				return ListPlayersResponse{}, NoActiveHeaderError{}
			}
			s = inActive
			continue

		case inActive:
			active, err := activePlayer(line)
			if err != nil {
				if line == disconnectedHeader {
					s = inDisconnected
					continue
				}
				return ListPlayersResponse{}, err
			}
			activePlayers = append(activePlayers, &active)

		case inDisconnected:
			disconnected, err := disconnectedPlayer(line)
			if err != nil {
				return ListPlayersResponse{}, err
			}
			disconnectedPlayers = append(disconnectedPlayers, &disconnected)
		}
	}

	return ListPlayersResponse{
		Active:       activePlayers,
		Disconnected: disconnectedPlayers,
	}, nil
}

func activePlayer(line string) (ActivePlayer, error) {
	var record activePlayerRecord

	err := activePlayerRe.MatchToTarget(line, &record)
	if err != nil {
		return ActivePlayer{}, LineNotMatchedError{line: line}
	}
	return intoActivePlayer(record), nil
}

func intoActivePlayer(record activePlayerRecord) ActivePlayer {
	var teamID *int32
	if record.TeamID != "N/A" {
		parsed, _ := strconv.ParseInt(record.TeamID, 10, 32)
		teamID = new(int32(parsed))
	}

	var squadID *int32
	if record.SquadID != "N/A" {
		parsed, _ := strconv.ParseInt(record.SquadID, 10, 32)
		squadID = new(int32(parsed))
	}

	var steamID *string
	if record.WithSteam {
		steamID = &record.SteamID
	}

	active := ActivePlayer{
		ConnectID:   record.ID,
		EosID:       record.EosID,
		SteamID:     steamID,
		Name:        record.Name,
		TeamNumber:  teamID,
		SquadNumber: squadID,
		IsLeader:    record.IsLeader,
		Role:        record.Role,
	}

	return active
}

func disconnectedPlayer(line string) (DisconnectedPlayer, error) {
	var record disconnectedPlayerRecord
	err := disconnectedPlayerRe.MatchToTarget(line, &record)
	if err != nil {
		return DisconnectedPlayer{}, LineNotMatchedError{}
	}

	return intoDisconnectedPlayer(record), nil
}

func intoDisconnectedPlayer(record disconnectedPlayerRecord) DisconnectedPlayer {
	var steamID *string
	if record.WithSteam {
		steamID = &record.SteamID
	}

	return DisconnectedPlayer{
		ConnectID:        record.ConnectID,
		EosID:            record.EosID,
		SteamID:          steamID,
		Name:             record.Name,
		DisconnectedSecs: record.Min*60 + record.Sec,
	}
}

type (
	InvalidActivePlayersHeaderError struct{ line string }
	InvalidDisconnectedLineError    struct{ line string }
	LineNotMatchedError             struct{ line string }
	NoActiveHeaderError             struct{}
	NoDisconnectedHeaderError       struct{}
	UnexpectedEOFError              struct{}
)

func (NoActiveHeaderError) Error() string {
	return "missing header for active players"
}

func (err InvalidActivePlayersHeaderError) Error() string {
	return fmt.Sprintf("invalid header for active players: `%v`", err.line)
}

func (NoDisconnectedHeaderError) Error() string {
	return "missing header for disconnected players"
}

func (err InvalidDisconnectedLineError) Error() string {
	return fmt.Sprintf("invalid disconnected player line: `%v`", err.line)
}

func (UnexpectedEOFError) Error() string {
	return "expected more data"
}

func (err LineNotMatchedError) Error() string {
	return fmt.Sprintf("failed to match line: `%v`", err.line)
}
