// SPDX-FileCopyrightText: 2026 koitra <koitra@rhoti.com>
//
// SPDX-License-Identifier: SSPL-1.0

package squadrcon

import (
	"bufio"
	"errors"
	"fmt"
)

type emptyPacket struct{}

func readEmptyPacket(src *bufio.Reader) (emptyPacket, error) {
	ofsPacket := [7]byte{0, 1, 0, 0, 0, 0, 0}
	const packetLen = len(ofsPacket)
	bytes, err := src.Peek(packetLen)
	if err != nil {
		return emptyPacket{}, err
	}
	for idx, byte := range bytes {
		if byte != ofsPacket[idx] {
			return emptyPacket{}, errNotEmptyPacket
		}
	}

	_, err = src.Discard(packetLen)
	if err != nil {
		return emptyPacket{}, fmt.Errorf("failed to discard bytes: %w", err)
	}

	return emptyPacket{}, nil
}

var errNotEmptyPacket = errors.New("not an empty packet")
