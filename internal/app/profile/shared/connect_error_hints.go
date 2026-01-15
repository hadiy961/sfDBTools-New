// File : internal/app/profile/shared/connect_error_hints.go
// Deskripsi : (DEPRECATED) Facade hint error koneksi profile
// Author : Hadiyatna Muflihun
// Tanggal : 9 Januari 2026
// Last Modified : 14 Januari 2026

package shared

import (
	profileconn "sfdbtools/internal/app/profile/connection"
)

type ConnectErrorKind = profileconn.ConnectErrorKind

const (
	ConnectErrorKindSSH = profileconn.ConnectErrorKindSSH
	ConnectErrorKindDB  = profileconn.ConnectErrorKindDB
)

type ConnectErrorInfo = profileconn.ConnectErrorInfo

func DescribeConnectError(err error) ConnectErrorInfo {
	return profileconn.DescribeConnectError(err)
}
