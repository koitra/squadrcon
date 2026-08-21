// SPDX-FileCopyrightText: 2026 koitra <koitra@rhoti.com>
//
// SPDX-License-Identifier: SSPL-1.0

package commands

import (
	"fmt"
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

func TestListSquadsCreatorIDs(t *testing.T) {
	tests := []struct {
		name   string
		ids    string
		expect Squad
	}{
		{
			name: "steam only",
			ids:  "EOS: 33333333333333333333333333333333 steam: 333333333333",
			expect: Squad{
				Number:         5,
				Name:           "Squad 5",
				Size:           1,
				Locked:         true,
				CreatorName:    "Name3",
				CreatorEosID:   "33333333333333333333333333333333",
				CreatorSteamID: new("333333333333"),
			},
		},
		{
			name: "epic only",
			ids:  "EOS: 44444444444444444444444444444444 epic: EpicAuthId.42",
			expect: Squad{
				Number:        5,
				Name:          "Squad 5",
				Size:          1,
				Locked:        true,
				CreatorName:   "Name3",
				CreatorEosID:  "44444444444444444444444444444444",
				CreatorEpicID: new("EpicAuthId.42"),
			},
		},
		{
			name: "all platforms",
			ids:  "EOS: 55555555555555555555555555555555 steam: 5555555555 epic: SomeEpicId",
			expect: Squad{
				Number:         5,
				Name:           "Squad 5",
				Size:           1,
				Locked:         true,
				CreatorName:    "Name3",
				CreatorEosID:   "55555555555555555555555555555555",
				CreatorSteamID: new("5555555555"),
				CreatorEpicID:  new("SomeEpicId"),
			},
		},
		{
			name: "eos only",
			ids:  "EOS: 66666666666666666666666666666666",
			expect: Squad{
				Number:       5,
				Name:         "Squad 5",
				Size:         1,
				Locked:       true,
				CreatorName:  "Name3",
				CreatorEosID: "66666666666666666666666666666666",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := fmt.Sprintf(`----- Active Squads -----
Team ID: 1 (Lord Strathcona's Horse Regiment)
ID: 5 | Name: Squad 5 | Size: 1 | Locked: True | Creator Name: Name3 | Creator Online IDs: %s`, tt.ids)

			expect := ListSquadsResponse{
				Teams: []*Team{
					{
						TeamID:      1,
						FactionName: "Lord Strathcona's Horse Regiment",
						Squads:      []*Squad{&tt.expect},
					},
				},
			}
			result, err := parseListSquads(body)
			require.NoError(t, err)
			require.Equal(t, expect, result)
		})
	}
}

func TestListSquadsWithoutEosCreatorID(t *testing.T) {
	body := `----- Active Squads -----
Team ID: 1 (Lord Strathcona's Horse Regiment)
ID: 5 | Name: Squad 5 | Size: 1 | Locked: True | Creator Name: Name3 | Creator Online IDs: steam: 333333333333`
	_, err := parseListSquads(body)
	require.ErrorContains(t, err, "no EOS ID")
}
