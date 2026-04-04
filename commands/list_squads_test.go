// SPDX-FileCopyrightText: 2026 koitra <koitra@rhoti.com>
//
// SPDX-License-Identifier: SSPL-1.0

package commands

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestListSquads(t *testing.T) {
	body := `----- Active Squads -----
Team ID: 1 (Lord Strathcona's Horse Regiment)
Team ID: 2 (118th Combined Arms Brigade)
ID: 5 | Name: Squad 5 | Size: 1 | Locked: True | Creator Name: Name3 | Creator Online IDs: EOS: 33333333333333333333333333333333 steam: 333333333333
ID: 2 | Name: Squad 2 | Size: 1 | Locked: True | Creator Name: Name3 | Creator Online IDs: EOS: 33333333333333333333333333333333 steam: 333333333333`
	steamID := "333333333333"
	expect := ListSquadsResponse{
		Teams: []*Team{
			{
				TeamID:      1,
				FactionName: "Lord Strathcona's Horse Regiment",
				Squads:      []*Squad{},
			},
			{
				TeamID:      2,
				FactionName: "118th Combined Arms Brigade",
				Squads: []*Squad{
					{
						Number:         5,
						Name:           "Squad 5",
						Size:           1,
						Locked:         true,
						CreatorName:    "Name3",
						CreatorEosID:   "33333333333333333333333333333333",
						CreatorSteamID: &steamID,
					},
					{
						Number:         2,
						Name:           "Squad 2",
						Size:           1,
						Locked:         true,
						CreatorName:    "Name3",
						CreatorEosID:   "33333333333333333333333333333333",
						CreatorSteamID: &steamID,
					},
				},
			},
		},
	}
	result, err := parseListSquads(body)
	require.NoError(t, err)
	require.Equal(t, expect, result)
}

func TestListSquadsEmpty(t *testing.T) {
	body := "----- Active Squads -----"
	expect := ListSquadsResponse{
		Teams: []*Team{},
	}
	result, err := parseListSquads(body)
	require.NoError(t, err)
	require.Equal(t, result, expect)
}
