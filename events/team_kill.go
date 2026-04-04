// SPDX-FileCopyrightText: 2026 koitra <koitra@rhoti.com>
//
// SPDX-License-Identifier: SSPL-1.0

package events

import (
	"fmt"

	"github.com/oriser/regroup"
)

type TeamKill struct {
	KillerName string
	VictimName string
}

var teamKillRe = regroup.MustCompile(
	`\[ChatAdmin\] ASQKillDeathRuleset : Player\s+(?P<killer>.+) Team Killed Player\s+(?P<victim>.+)`,
)

func (e *TeamKill) parse(b []byte) error {
	type record struct {
		Killer string `regroup:"killer"`
		Victim string `regroup:"victim"`
	}

	var r record
	err := teamKillRe.MatchToTarget(string(b), &r)
	if err != nil {
		return fmt.Errorf("match failed: %w", err)
	}

	e.KillerName = r.Killer
	e.VictimName = r.Victim

	return nil
}
