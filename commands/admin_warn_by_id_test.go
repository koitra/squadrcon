// SPDX-FileCopyrightText: 2026 koitra <koitra@rhoti.com>
//
// SPDX-License-Identifier: SSPL-1.0

package commands

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAdminWarnSuccess(t *testing.T) {
	body := `Remote admin has warned player [CLAN]  Name. Message was "MESSAGE HERE"`
	expect := AdminWarnByIDResponse{
		ID:      "[CLAN]  Name",
		Message: "MESSAGE HERE",
	}

	res, err := parseAdminWarn(body)
	require.NoError(t, err)

	require.Equal(t, expect, res)
}

func TestAdminWarnFailed(t *testing.T) {
	body := "Could not find player 72221128082822272"
	expect := WarnedPlayerNotFoundError{
		ID: "72221128082822272",
	}
	_, err := parseAdminWarn(body)
	require.EqualError(t, err, expect.Error())
}
