// SPDX-FileCopyrightText: 2026 koitra <koitra@rhoti.com>
//
// SPDX-License-Identifier: SSPL-1.0

package commands

import (
	"fmt"

	"github.com/oriser/regroup"
)

type AdminWarnByID struct {
	EosID string
	// Must not be empty
	// TODO: validate
	Message string
}

func (c AdminWarnByID) ToBody() string {
	return fmt.Sprintf("AdminWarn %v %v", c.EosID, c.Message)
}

func (AdminWarnByID) ParseBody(body string) (AdminWarnByIDResponse, error) {
	return parseAdminWarn(body)
}

type (
	AdminWarnByIDResponse struct {
		ID      string
		Message string
	}

	adminWarnRecord struct {
		Player  string `regroup:"player"`
		Message string `regroup:"message"`
	}
)

func parseAdminWarn(body string) (AdminWarnByIDResponse, error) {
	var rec adminWarnRecord
	err := adminWarnRecordRe.MatchToTarget(body, &rec)
	if err != nil {
		var notFound WarnedPlayerNotFoundError
		err := adminWarnFailedRe.MatchToTarget(body, &notFound)
		if err != nil {
			return AdminWarnByIDResponse{}, fmt.Errorf("unknown body: `%v`", body)
		}
		return AdminWarnByIDResponse{}, notFound
	}

	res := AdminWarnByIDResponse{
		ID:      rec.Player,
		Message: rec.Message,
	}
	return res, nil
}

type WarnedPlayerNotFoundError struct {
	ID string `regroup:"player"`
}

func (err WarnedPlayerNotFoundError) Error() string {
	return fmt.Sprintf("player `%v` was not found", err.ID)
}

var (
	adminWarnRecordRe = regroup.MustCompile(
		`Remote admin has warned player (?P<player>.*)\. Message was "(?P<message>.*)"$`,
	)
	adminWarnFailedRe = regroup.MustCompile(`Could not find player (?P<player>.*)$`)
)
