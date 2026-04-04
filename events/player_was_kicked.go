// SPDX-FileCopyrightText: 2026 koitra <koitra@rhoti.com>
//
// SPDX-License-Identifier: SSPL-1.0

package events

import (
	"fmt"

	"github.com/oriser/regroup"
)

type PlayerWasKicked struct {
	PlayerName string
	Reason     string
}

var playerWasKickedRe = regroup.MustCompile(
	`(?s)\s*(?P<playerName>.+) was kicked: (?P<reason>.+)$`,
)

func (e *PlayerWasKicked) parse(b []byte) error {
	type record struct {
		PlayerName string `regroup:"playerName"`
		Reason     string `regroup:"reason"`
	}

	var r record
	err := playerWasKickedRe.MatchToTarget(string(b), &r)
	if err != nil {
		return fmt.Errorf("match failed: %w", err)
	}

	e.PlayerName = r.PlayerName
	e.Reason = r.Reason

	return nil
}
