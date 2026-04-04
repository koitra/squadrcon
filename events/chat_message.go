// SPDX-FileCopyrightText: 2026 koitra <koitra@rhoti.com>
//
// SPDX-License-Identifier: SSPL-1.0

package events

import (
	"fmt"

	"github.com/oriser/regroup"
)

type ChatScope int32

const (
	ChatScopeAll ChatScope = iota + 1
	ChatScopeTeam
	ChatScopeSquad
	ChatScopeBroadcast
	ChatScopeAdmin
)

type ChatMessage struct {
	PlayerName string
	EosID      string
	SteamID    string
	Message    string
	Scope      ChatScope
}

var chatMessageRe = regroup.MustCompile(
	`\[Chat(?P<scope>.+)\] \[Online IDs:EOS: (?P<eosID>.*) steam: (?P<steamID>\d+)\]\s+(?P<playerName>.+) : (?P<message>.+)`,
)

func (e *ChatMessage) parse(b []byte) error {
	type record struct {
		EosID      string `regroup:"eosID"`
		SteamID    string `regroup:"steamID"`
		PlayerName string `regroup:"playerName"`
		Message    string `regroup:"message"`
		Scope      string `regroup:"scope"`
	}

	var r record
	err := chatMessageRe.MatchToTarget(string(b), &r)
	if err != nil {
		return fmt.Errorf("match failed: %w", err)
	}

	e.EosID = r.EosID
	e.SteamID = r.SteamID
	e.PlayerName = r.PlayerName
	e.Message = r.Message

	switch r.Scope {
	case "All":
		e.Scope = ChatScopeAll
	case "Team":
		e.Scope = ChatScopeTeam
	case "Squad":
		e.Scope = ChatScopeSquad
	case "Broadcast":
		e.Scope = ChatScopeBroadcast
	case "Admin":
		e.Scope = ChatScopeAdmin
	default:
		return fmt.Errorf("unknown chat scope: %v", r.Scope)
	}

	return nil
}
