// SPDX-FileCopyrightText: 2026 koitra <koitra@rhoti.com>
//
// SPDX-License-Identifier: SSPL-1.0

package commands

type (
	RconCommand[Response any] interface {
		ToBody() string
		ParseBody(string) (Response, error)
	}
)
