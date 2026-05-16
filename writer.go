// SPDX-FileCopyrightText: 2026 koitra <koitra@rhoti.com>
//
// SPDX-License-Identifier: SSPL-1.0

package squadrcon

import (
	"context"
	"fmt"
	"io"
)

type writer struct {
	reqs <-chan request
	dst  io.Writer
	msgs chan<- connectionMessage
}

func (w *writer) run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case req := <-w.reqs:
			if err := req.write(w.dst); err != nil {
				w.msgs <- writerError{Err: err}
				return
			}
		}
	}
}

type writerError struct {
	Err error
}

func (e writerError) Error() string {
	return fmt.Sprintf("writer error: %v", e.Err)
}

func (writerError) connectionMessage() {}
