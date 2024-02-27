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

func NewNotImplemented(msg string) ObError {
	return &httpErr{
		errorType: ErrNotImplemented,
		message:   msg,
	}
}

func NewBadRequest(msg string) ObError {
	return &httpErr{
		errorType: ErrBadRequest,
		message:   msg,
	}
}

func NewUnauthorized(msg string) ObError {
	return &httpErr{
		errorType: ErrUnauthorized,
		message:   msg,
	}
}

func NewNotFound(msg string) ObError {
	return &httpErr{
		errorType: ErrNotFound,
		message:   msg,
	}
}

func NewInternal(msg string) ObError {
	return &httpErr{
		errorType: ErrInternal,
		message:   msg,
	}
}
