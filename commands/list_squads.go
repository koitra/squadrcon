// SPDX-FileCopyrightText: 2026 koitra <koitra@rhoti.com>
//
// SPDX-License-Identifier: SSPL-1.0

package commands

import (
	"fmt"
	"strings"

	"github.com/oriser/regroup"
)

type ListSquads struct{}

func (ListSquads) ToBody() string { return "ListSquads" }
func (ListSquads) ParseBody(body string) (ListSquadsResponse, error) {
	return parseListSquads(body)
}

type (
	ListSquadsResponse struct {
		Teams []*Team
	}

	Team struct {
		TeamID      int32
		FactionName string
		Squads      []*Squad
	}
	teamRecord struct {
		TeamID      int32  `regroup:"team_id"`
		FactionName string `regroup:"faction_name"`
	}

	Squad struct {
		Number         int32
		Name           string
		Size           int32
		Locked         bool
		CreatorName    string
		CreatorEosID   string
		CreatorSteamID *string
	}

	squadRecord struct {
		Number         int32  `regroup:"squad_id"`
		Name           string `regroup:"squad_name"`
		Size           int32  `regroup:"squad_size"`
		Locked         bool   `regroup:"locked"`
		CreatorName    string `regroup:"creator_name"`
		CreatorEosID   string `regroup:"eos_id"`
		CreatorSteamID string `regroup:"steam_id"`
		WithSteam      bool   `regroup:"steam_id,exists"`
	}
)

var (
	teamRe  = regroup.MustCompile(`Team ID: (?P<team_id>\d+) \((?P<faction_name>.*)\)$`)
	squadRe = regroup.MustCompile(
		`ID: (?P<squad_id>\d+) \| Name: (?P<squad_name>.*) \| Size: (?P<squad_size>\d+) \| Locked: (?P<locked>(True)|(False)) \| Creator Name: (?P<creator_name>.*) \| Creator Online IDs: EOS: (?P<eos_id>[\da-z]{32})( steam: (?P<steam_id>\d+))?$`,
	)
)

const listSquadsHeader = "----- Active Squads -----"

func parseListSquads(body string) (ListSquadsResponse, error) {
	normalized := strings.ReplaceAll(body, "\r\n", "\n")
	normalized = strings.ReplaceAll(normalized, "\r", "\n")
	lines := strings.Split(normalized, "\n")
	if len(lines) < 1 {
		return ListSquadsResponse{}, fmt.Errorf("invalid body: `%v`", body)
	}

	header := strings.TrimSpace(lines[0])

	if header != listSquadsHeader {
		return ListSquadsResponse{}, ErrInvalidHeader{Got: header}
	}
	teams := []*Team{}

	for idx := 1; idx < len(lines); {
		var team teamRecord
		err := teamRe.MatchToTarget(lines[idx], &team)
		if err != nil {
			return ListSquadsResponse{}, fmt.Errorf(
				"line `%v` is not a team record: %w",
				lines[idx],
				err,
			)
		}
		idx += 1

		squads := []*Squad{}

		for idx < len(lines) {
			var squad squadRecord
			err := squadRe.MatchToTarget(lines[idx], &squad)
			if err != nil {
				break
			}

			var steamId *string
			if squad.WithSteam {
				steamId = &squad.CreatorSteamID
			}
			squads = append(squads, &Squad{
				Number:         squad.Number,
				Name:           squad.Name,
				Size:           squad.Size,
				Locked:         squad.Locked,
				CreatorName:    squad.CreatorName,
				CreatorEosID:   squad.CreatorEosID,
				CreatorSteamID: steamId,
			})

			idx += 1
		}

		teams = append(teams, &Team{
			TeamID:      team.TeamID,
			FactionName: team.FactionName,
			Squads:      squads,
		})
	}

	return ListSquadsResponse{
		Teams: teams,
	}, nil
}

type (
	ErrInvalidHeader struct {
		Got string
	}
)

func (e ErrInvalidHeader) Error() string {
	return fmt.Sprintf("invalid header, got `%v`, expected `%v`", e.Got, listSquadsHeader)
}
