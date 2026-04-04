// SPDX-FileCopyrightText: 2026 koitra <koitra@rhoti.com>
//
// SPDX-License-Identifier: SSPL-1.0

package squadrcon

type ConnectionStatus int8

const (
	StatusHealthy  ConnectionStatus = 1
	StatusDegraded ConnectionStatus = 2
)
