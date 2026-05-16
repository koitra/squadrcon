// SPDX-FileCopyrightText: 2026 koitra <koitra@rhoti.com>
//
// SPDX-License-Identifier: SSPL-1.0

package squadrcon

import (
	"bufio"
	"context"
	"fmt"
)

type reader struct {
	src  *bufio.Reader
	msgs chan<- connectionMessage
}

func (r *reader) nun(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			if err := r.read(); err != nil {
				r.msgs <- readerError{Err: err}
				return
			}
		}
	}
}

func (r *reader) read() error {
	var err error

	empty, err := readEmptyPacket(r.src)
	if err != nil && err != errNotEmptyPacket {
		return err
	}

	if err == nil {
		r.msgs <- empty
		return nil
	}

	packet, err := readPacket(r.src)
	if err != nil {
		return err
	}

	r.msgs <- packet

	return nil
}

type readerError struct {
	Err error
}

func (e readerError) Error() string {
	return fmt.Sprintf("reader error: %v", e.Err)
}

func (readerError) connectionMessage() {}
