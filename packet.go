// SPDX-FileCopyrightText: 2026 koitra <koitra@rhoti.com>
//
// SPDX-License-Identifier: SSPL-1.0

package squadrcon

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

type packet struct {
	id   int32
	ty   int32
	body []byte
}

func readPacket(src io.Reader) (packet, error) {
	sizeBuf := make([]byte, 4)
	_, err := io.ReadFull(src, sizeBuf)
	if err != nil {
		return packet{}, fmt.Errorf("failed to read packet size: %w", err)
	}
	size := int32(binary.LittleEndian.Uint32(sizeBuf))

	if size < packetOverhead {
		return packet{}, errInvalidSize{Size: size}
	}

	packetBuf := make([]byte, size)
	_, err = io.ReadFull(src, packetBuf)
	if err != nil {
		return packet{}, err
	}

	packetReader := bytes.NewReader(packetBuf)
	var id int32
	err = binary.Read(packetReader, binary.LittleEndian, &id)
	if err != nil {
		return packet{}, err
	}

	var ty int32
	err = binary.Read(packetReader, binary.LittleEndian, &ty)
	if err != nil {
		return packet{}, err
	}

	body := make([]byte, size-packetOverhead)
	_, err = io.ReadFull(packetReader, body)
	if err != nil {
		return packet{}, fmt.Errorf("failed to read body: %w", err)
	}

	terminators := make([]byte, 2)
	_, err = io.ReadFull(packetReader, terminators)
	if err != nil {
		return packet{}, fmt.Errorf("failed to read terminators: %w", err)
	}

	bodyTerminator := terminators[0]
	if bodyTerminator != 0 {
		return packet{}, errInvalidBodyTerminator{Terminator: bodyTerminator}
	}

	packetTerminator := terminators[1]
	if packetTerminator != 0 {
		return packet{}, errInavlidPacketTerminator{Terminator: packetTerminator}
	}

	return packet{
		id:   id,
		ty:   ty,
		body: body,
	}, nil
}

func (p *packet) write(w io.Writer) error {
	err := binary.Write(w, binary.LittleEndian, int32(len(p.body))+packetOverhead)
	if err != nil {
		return err
	}

	err = binary.Write(w, binary.LittleEndian, &p.id)
	if err != nil {
		return err
	}
	err = binary.Write(w, binary.LittleEndian, &p.ty)
	if err != nil {
		return err
	}

	_, err = io.Copy(w, bytes.NewReader(p.body))
	if err != nil {
		return err
	}

	const terminator = byte(0)
	err = binary.Write(w, binary.LittleEndian, terminator)
	if err != nil {
		return err
	}

	err = binary.Write(w, binary.LittleEndian, terminator)
	if err != nil {
		return err
	}

	return nil
}

func (p *packet) IsEvent() bool {
	return p.id == 0 && p.ty == serverDataEventTy
}

type (
	errInvalidSize             struct{ Size int32 }
	errInvalidBodyTerminator   struct{ Terminator byte }
	errInavlidPacketTerminator struct{ Terminator byte }
)

func (e errInvalidSize) Error() string {
	return fmt.Sprintf("invalid packet size: %v", e.Size)
}

func (e errInvalidBodyTerminator) Error() string {
	return fmt.Sprintf("invalid body terminator: %v", e.Terminator)
}

func (e errInavlidPacketTerminator) Error() string {
	return fmt.Sprintf("invalid packet terminator: %v", e.Terminator)
}

func (packet) connectionMessage() {}
