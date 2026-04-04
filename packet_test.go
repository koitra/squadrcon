// SPDX-FileCopyrightText: 2026 koitra <koitra@rhoti.com>
//
// SPDX-License-Identifier: SSPL-1.0

package squadrcon

import (
	"bufio"
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadPacket(t *testing.T) {
	body := []byte("body")
	src := []byte{14, 0, 0, 0, 2, 0, 0, 0, 3, 0, 0, 0, 98, 111, 100, 121, 0, 0, 14}

	expect := packet{
		id:   2,
		ty:   serverDataAuthTy,
		body: body,
	}
	end := []byte{14}

	packetReader := bufio.NewReader(bytes.NewReader(src))
	result, err := readPacket(packetReader)
	assert.NoError(t, err)
	require.Equal(t, expect, result)

	leftOver, _ := io.ReadAll(packetReader)
	require.Equal(t, end, leftOver)
}

func TestPacketWrite(t *testing.T) {
	body := []byte("body")
	expect := []byte{14, 0, 0, 0, 1, 0, 0, 0, 1, 0, 0, 0, 98, 111, 100, 121, 0, 0}

	var packetID int32 = 1
	var packetTy int32 = 1

	packet := &packet{
		id:   packetID,
		ty:   packetTy,
		body: body,
	}

	dst := &bytes.Buffer{}

	err := packet.write(dst)
	require.NoError(t, err)

	written := dst.Bytes()
	require.Equal(t, expect, written)
}
