// SPDX-FileCopyrightText: 2026 koitra <koitra@rhoti.com>
//
// SPDX-License-Identifier: SSPL-1.0

package events

import (
	"fmt"

	"github.com/oriser/regroup"
)

type CouldNotFindPlayer struct {
	Criterion string
}

var couldNotFindRe = regroup.MustCompile(
	`Could not find player (?P<criterion>\d+)`,
)

func (e *CouldNotFindPlayer) parse(b []byte) error {
	type record struct {
		Criterion string `regroup:"criterion"`
	}

	var r record
	err := couldNotFindRe.MatchToTarget(string(b), &r)
	if err != nil {
		return fmt.Errorf("match failed: %w", err)
	}
	e.Criterion = r.Criterion

	return nil
}
