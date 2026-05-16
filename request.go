// SPDX-FileCopyrightText: 2026 koitra <koitra@rhoti.com>
//
// SPDX-License-Identifier: SSPL-1.0

package squadrcon

import (
	"io"
	"iter"
)

type request struct {
	packets []packet
}

func newRequest(
	id int32,
	ty int32,
	body []byte,
) request {
	packets := []packet{}

	for bodyChunk := range chunks(body, packetBodyMaxSize) {
		write := packet{
			id:   id,
			ty:   ty,
			body: bodyChunk,
		}

		packets = append(packets, write)
	}

	return request{
		packets,
	}
}

func (r *request) write(w io.Writer) error {
	for _, packet := range r.packets {
		err := packet.write(w)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *request) appendEmpty() {
	r.packets = append(r.packets, packet{id: r.packets[0].id, ty: serverDataResponseValueTy})
}

func chunks(data []byte, chunkSize int) iter.Seq[[]byte] {
	return func(yield func([]byte) bool) {
		for i := 0; i < len(data); i += chunkSize {
			end := min(i+chunkSize, len(data))

			if !yield(data[i:end]) {
				return
			}
		}
	}
}
