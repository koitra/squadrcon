// SPDX-FileCopyrightText: 2026 koitra <koitra@rhoti.com>
//
// SPDX-License-Identifier: SSPL-1.0

package squadrcon

const (
	packetOverhead    = 10
	packetBodyMaxSize = 4096 - packetOverhead
)

const (
	serverDataAuthTy          = 3
	serverDataAuthResponseTy  = 2
	serverDataExecCommandTy   = 2
	serverDataResponseValueTy = 0
	serverDataEventTy         = 1
)
