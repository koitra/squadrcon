// SPDX-FileCopyrightText: 2026 koitra <koitra@rhoti.com>
//
// SPDX-License-Identifier: SSPL-1.0

package commands

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestActivePlayer(t *testing.T) {
	tests := []struct {
		name   string
		line   string
		expect ActivePlayer
	}{
		{
			name: "with squad",
			line: "ID: 27 | Online IDs: EOS: 02223fa0faf147ac93aa3da72f232b20 steam: 72211442292221100 | Name:  受狂  S1100 | Team ID: 1 | Squad ID: 7 | Is Leader: True | Role: ADF_Medic_02",
			expect: ActivePlayer{
				ConnectID:   27,
				EosID:       "02223fa0faf147ac93aa3da72f232b20",
				SteamID:     new("72211442292221100"),
				Name:        "受狂  S1100",
				TeamNumber:  new(int32(1)),
				SquadNumber: new(int32(7)),
				IsLeader:    true,
				Role:        "ADF_Medic_02",
			},
		},
		{
			name: "without team",
			line: "ID: 27 | Online IDs: EOS: 02223fa0faf147ac93aa3da72f232b20 steam: 72211442292221100 | Name:  受狂  S1100 | Team ID: 1 | Squad ID: 7 | Is Leader: True | Role: ADF_Medic_02",
			expect: ActivePlayer{
				ConnectID:   27,
				EosID:       "02223fa0faf147ac93aa3da72f232b20",
				SteamID:     new("72211442292221100"),
				Name:        "受狂  S1100",
				TeamNumber:  new(int32(1)),
				SquadNumber: new(int32(7)),
				IsLeader:    true,
				Role:        "ADF_Medic_02",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := activePlayer(tt.line)
			require.NoError(t, err)
			require.Equal(t, tt.expect, result)
		})
	}
}

func TestDisconnectedPlayer(t *testing.T) {
	line := "ID: 38 | Online IDs: EOS: 02223fa0faf147ac93aa3da72f232b20 steam: 72211442292221100 | Since Disconnect: 04m.10s | Name:  pehf))"
	expect := DisconnectedPlayer{
		ConnectID:        38,
		EosID:            "02223fa0faf147ac93aa3da72f232b20",
		SteamID:          new("72211442292221100"),
		Name:             "pehf))",
		DisconnectedSecs: 250,
	}
	result, err := disconnectedPlayer(line)
	require.NoError(t, err)
	require.Equal(t, expect, result)
}

func TestListPlayersWithoutDisconnected(t *testing.T) {
	body := `----- Active Players -----
ID: 108 | Online IDs: EOS: 0002a1fc90d241fcbb83baa13c59df84 steam: 00000000000000000 | Name: clan nickame | Team ID: 2 | Squad ID: 8 | Is Leader: True | Role: RGF_SL_01
ID: 103 | Online IDs: EOS: 0002e129bb484f5894248fbb91f38a26 steam: 00000000000000000 | Name:  克crying克 | Team ID: 2 | Squad ID: N/A | Is Leader: False | Role: RGF_LAT_02`

	expect := ListPlayersResponse{
		Active: []*ActivePlayer{
			{
				ConnectID:   108,
				EosID:       "0002a1fc90d241fcbb83baa13c59df84",
				SteamID:     new("00000000000000000"),
				Name:        "clan nickame",
				TeamNumber:  new(int32(2)),
				SquadNumber: new(int32(8)),
				IsLeader:    true,
				Role:        "RGF_SL_01",
			},
			{
				ConnectID:   103,
				EosID:       "0002e129bb484f5894248fbb91f38a26",
				SteamID:     new("00000000000000000"),
				Name:        "克crying克",
				TeamNumber:  new(int32(2)),
				SquadNumber: nil,
				IsLeader:    false,
				Role:        "RGF_LAT_02",
			},
		},
	}
	result, err := parseListPlayers(body)

	require.NoError(t, err)
	require.Equal(t, expect, result)
}

func TestListPlayers(t *testing.T) {
	body := `----- Active Players -----
ID: 108 | Online IDs: EOS: 0002a1fc90d241fcbb83baa13c59df84 steam: 00000000000000000 | Name: clan nickame | Team ID: 2 | Squad ID: 8 | Is Leader: True | Role: RGF_SL_01
ID: 103 | Online IDs: EOS: 0002e129bb484f5894248fbb91f38a26 steam: 00000000000000000 | Name:  克crying克 | Team ID: 2 | Squad ID: N/A | Is Leader: False | Role: RGF_LAT_02
----- Recently Disconnected Players [Max of 15] -----
ID: 95 | Online IDs: EOS: 0002321bea6c4dc49ba18c4c5bd0aefd steam: 00000000000000000 | Since Disconnect: 02m.50s | Name:  Trex
ID: 86 | Online IDs: EOS: 0002222205da2e22ll233333cd4f8e2f | Since Disconnect: 04m.51s | Name:  kfke
ID: 36 | Online IDs: EOS: 00023695959a49d28315e5300a651bb7 steam: 00000000000000000 | Since Disconnect: 01m.09s | Name:  EsToNiAn444`
	expect := ListPlayersResponse{
		Active: []*ActivePlayer{
			{
				ConnectID:   108,
				EosID:       "0002a1fc90d241fcbb83baa13c59df84",
				SteamID:     new("00000000000000000"),
				Name:        "clan nickame",
				TeamNumber:  new(int32(2)),
				SquadNumber: new(int32(8)),
				IsLeader:    true,
				Role:        "RGF_SL_01",
			},
			{
				ConnectID:   103,
				EosID:       "0002e129bb484f5894248fbb91f38a26",
				SteamID:     new("00000000000000000"),
				Name:        "克crying克",
				TeamNumber:  new(int32(2)),
				SquadNumber: nil,
				IsLeader:    false,
				Role:        "RGF_LAT_02",
			},
		},
		Disconnected: []*DisconnectedPlayer{
			{
				ConnectID:        95,
				EosID:            "0002321bea6c4dc49ba18c4c5bd0aefd",
				SteamID:          new("00000000000000000"),
				Name:             "Trex",
				DisconnectedSecs: 170,
			},
			{
				ConnectID:        86,
				EosID:            "0002222205da2e22ll233333cd4f8e2f",
				DisconnectedSecs: 291,
				Name:             "kfke",
			},
			{
				ConnectID:        36,
				EosID:            "00023695959a49d28315e5300a651bb7",
				SteamID:          new("00000000000000000"),
				Name:             "EsToNiAn444",
				DisconnectedSecs: 69,
			},
		},
	}

	result, err := parseListPlayers(body)
	require.NoError(t, err)
	require.Equal(t, result, expect)
}

func TestActivePlayerWithoutSteamID(t *testing.T) {
	line := `ID: 0 | Online IDs: EOS: 0022245a5f394b2e22227e5222222229 | Name:  Nameyrt | Team ID: 1 | Squad ID: 3 | Is Leader: False | Role: WPMC_Rifleman_06`
	player, err := activePlayer(line)
	require.NoError(t, err)
	expect := ActivePlayer{
		ConnectID:   0,
		EosID:       "0022245a5f394b2e22227e5222222229",
		SteamID:     nil,
		Name:        "Nameyrt",
		TeamNumber:  new(int32(1)),
		SquadNumber: new(int32(3)),
		IsLeader:    false,
		Role:        "WPMC_Rifleman_06",
	}
	require.Equal(t, expect, player)
}
