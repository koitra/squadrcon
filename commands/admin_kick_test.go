// SPDX-FileCopyrightText: 2026 koitra <koitra@rhoti.com>
//
// SPDX-License-Identifier: SSPL-1.0

package commands

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAdminKick(t *testing.T) {
	body := `PlayerNick was kicked: REASON
Kicked player 55. [Online IDs= EOS: 0002222d9d1333333333c5333333331f steam: 22221134667047434] PlayerNick`
	expect := AdminKickResponse{
		ConnectID:  55,
		Message:    "REASON",
		EosID:      "0002222d9d1333333333c5333333331f",
		SteamID:    "22221134667047434",
		PlayerName: "PlayerNick",
	}

	result, err := ParseAdminKick(body)
	require.NoError(t, err)
	require.Equal(t, expect, result)
}

func TestAdminKickFailed(t *testing.T) {
	body := `ERROR: Unable to find player with name or id (71111112062222224)`
	expect := KickPlayerNotFoundError{
		ID: "71111112062222224",
	}

	_, err := ParseAdminKick(body)
	require.EqualError(t, err, expect.Error())
}
