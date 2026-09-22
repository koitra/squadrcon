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
	src *bufio.Reader
	bus chan<- any
}

func (r *reader) run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			if err := r.read(); err != nil {
				r.bus <- readError{err}
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
		r.bus <- empty
		return nil
	}

	packet, err := readPacket(r.src)
	if err != nil {
		return err
	}

	r.bus <- packet

	return nil
}

type readError struct {
	Err error
}

func (e readError) Error() string {
	return fmt.Sprintf("read: %v", e.Err)

}

func (e readError) Unwrap() error {
	return e.Err
}
