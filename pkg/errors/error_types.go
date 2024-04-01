/*
Copyright (c) 2023 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package errors

type ErrorType string

var (
	ErrNotFound           ErrorType = "NotFound"
	ErrUnauthorized       ErrorType = "Unauthorized"
	ErrForbidden          ErrorType = "Forbidden"
	ErrConflict           ErrorType = "Conflict"
	ErrInternal           ErrorType = "Internal"
	ErrTimeout            ErrorType = "Timeout"
	ErrTooManyRequests    ErrorType = "TooManyRequests"
	ErrBadRequest         ErrorType = "BadRequest"
	ErrInvalid            ErrorType = "Invalid"
	ErrNotSupported       ErrorType = "NotSupported"
	ErrAlreadyExists      ErrorType = "AlreadyExists"
	ErrNotReady           ErrorType = "NotReady"
	ErrNotImplemented     ErrorType = "NotImplemented"
	ErrServiceUnavailable ErrorType = "ServiceUnavailable"

	ErrInsufficientResource ErrorType = "InsufficientResource"
)
