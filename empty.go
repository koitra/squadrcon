// SPDX-FileCopyrightText: 2026 koitra <koitra@rhoti.com>
//
// SPDX-License-Identifier: SSPL-1.0

package squadrcon

import (
	"bufio"
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
			return emptyPacket{}, notAnEmptyPacketError{}
		}
	}

	_, err = src.Discard(packetLen)
	if err != nil {
		return emptyPacket{}, fmt.Errorf("failed to discard bytes: %w", err)
	}

	return emptyPacket{}, nil
}

type notAnEmptyPacketError struct{}

func (notAnEmptyPacketError) Error() string {
	return "not an empty packet"
}
