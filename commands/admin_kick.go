// SPDX-FileCopyrightText: 2026 koitra <koitra@rhoti.com>
//
// SPDX-License-Identifier: SSPL-1.0

package commands

import (
	"fmt"
	"strings"

	"github.com/oriser/regroup"
)

type AdminKick struct {
	// SteamID or nickname
	Player string
	// Must not be empty
	// TODO: validate
	Reason string
}

func (c AdminKick) ToBody() string {
	return fmt.Sprintf("AdminKick \"%v\" \"%v\"", c.Player, c.Reason)
}

func (AdminKick) ParseBody(body string) (AdminKickResponse, error) {
	return ParseAdminKick(body)
}

type (
	AdminKickResponse struct {
		ConnectID  int32
		EosID      string
		SteamID    string
		Message    string
		PlayerName string
	}

	adminKickRecord struct {
		ConnectID int32  `regroup:"connect_id"`
		EosID     string `regroup:"eos_id"`
		SteamID   string `regroup:"steam_id"`
		Message   string `regroup:"message"`
		Name      string `regroup:"name"`
	}

	KickPlayerNotFoundError struct {
		ID string `regroup:"id"`
	}
)

func ParseAdminKick(body string) (AdminKickResponse, error) {
	normalized := strings.ReplaceAll(body, "\r\n", "\n")
	normalized = strings.ReplaceAll(normalized, "\r", "\n")
	var record adminKickRecord
	err := adminKickRecordRe.MatchToTarget(normalized, &record)
	if err != nil {
		var failed KickPlayerNotFoundError
		err = adminKickFailedRE.MatchToTarget(normalized, &failed)
		if err != nil {
			return AdminKickResponse{}, fmt.Errorf("invalid body: `%v`\n%w", normalized, err)
		}
		return AdminKickResponse{}, failed
	}

	resp := AdminKickResponse{
		ConnectID:  record.ConnectID,
		EosID:      record.EosID,
		SteamID:    record.SteamID,
		Message:    record.Message,
		PlayerName: record.Name,
	}

	return resp, nil
}

func (err KickPlayerNotFoundError) Error() string {
	return fmt.Sprintf("player with `%v` was not found", err.ID)
}

var (
	adminKickRecordRe = regroup.MustCompile(
		`(?m)(.*) was kicked: (?P<message>.*)$\nKicked player (?P<connect_id>\d+)\. \[Online IDs= EOS: (?P<eos_id>[\da-z]{32}) steam: (?P<steam_id>\d+)\] (?P<name>.*)$`,
	)

	adminKickFailedRE = regroup.MustCompile(
		`ERROR: Unable to find player with name or id \((?P<id>.*)\)$`,
	)
)
