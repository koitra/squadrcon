// SPDX-FileCopyrightText: 2026 koitra <koitra@rhoti.com>
//
// SPDX-License-Identifier: SSPL-1.0

package events

import (
	"fmt"
)

type Event interface {
	parse([]byte) error
}

func ParseEvent(b []byte) (Event, error) {
	if len(b) == 0 {
		return nil, fmt.Errorf("empty event body")
	}

	var warnEvent Warn
	if err := warnEvent.parse(b); err == nil {
		return &warnEvent, nil
	}

	var teamKillEvent TeamKill
	if err := teamKillEvent.parse(b); err == nil {
		return &teamKillEvent, nil
	}

	var couldNotFindPlayerEvent CouldNotFindPlayer
	if err := couldNotFindPlayerEvent.parse(b); err == nil {
		return &couldNotFindPlayerEvent, nil
	}

	var chatMessageEvent ChatMessage
	if err := chatMessageEvent.parse(b); err == nil {
		return &chatMessageEvent, nil
	}

	var kickedPlayerEvent KickedPlayer
	if err := kickedPlayerEvent.parse(b); err == nil {
		return &kickedPlayerEvent, nil
	}

	var playerWasKickedEvent PlayerWasKicked
	if err := playerWasKickedEvent.parse(b); err == nil {
		return &playerWasKickedEvent, nil
	}

	var createdSquadEvent CreatedSquad
	if err := createdSquadEvent.parse(b); err == nil {
		return &createdSquadEvent, nil
	}

	return nil, UnknownEventError{Body: b}
}

type UnknownEventError struct {
	Body []byte
}

func (e UnknownEventError) Error() string {
	return fmt.Sprintf("unknown event: %v", string(e.Body))
}
