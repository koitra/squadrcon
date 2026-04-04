// SPDX-FileCopyrightText: 2026 koitra <koitra@rhoti.com>
//
// SPDX-License-Identifier: SSPL-1.0

package events

import (
	"fmt"

	"github.com/oriser/regroup"
)

type Warn struct {
	PlayerName string
	Message    string
}

var warnRe = regroup.MustCompile(
	`(?s)Remote admin has warned player\s+(?P<playerName>.+)\. Message was "(?P<message>.+)"`,
)

func (e *Warn) parse(b []byte) error {
	type record struct {
		PlayerName string `regroup:"playerName"`
		Message    string `regroup:"message"`
	}

	var r record
	err := warnRe.MatchToTarget(string(b), &r)
	if err != nil {
		return fmt.Errorf("match failed: %w", err)
	}

	e.PlayerName = r.PlayerName
	e.Message = r.Message

	return nil
}
