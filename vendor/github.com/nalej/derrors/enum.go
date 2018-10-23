/*
 * Copyright 2018 Daisho
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Definition of the supported types of error. Based on https://godoc.org/google.golang.org/grpc/codes

package derrors

type ErrorType int

const (
	Generic ErrorType = iota + 1
	Canceled
	InvalidArgument
	DeadlineExceeded
	NotFound
	AlreadyExists
	PermissionDenied
	ResourceExhausted
	FailedPrecondition
	Aborted
	OutOfRange
	Unimplemented
	Internal
	Unavailable
	Unauthenticated
)

var ErrorTypesValues = map[string]ErrorType{
	"Generic":            Generic,
	"Canceled":           Canceled,
	"InvalidArgument":    InvalidArgument,
	"DeadlineExceeded":   DeadlineExceeded,
	"NotFound":           NotFound,
	"AlreadyExists":      AlreadyExists,
	"PermissionDenied":   PermissionDenied,
	"ResourceExhausted":  ResourceExhausted,
	"FailedPrecondition": FailedPrecondition,
	"Aborted":            Aborted,
	"OutOfRange":         OutOfRange,
	"Unimplemented":      Unimplemented,
	"Internal":           Internal,
	"Unavailable":        Unavailable,
	"Unauthenticated":    Unauthenticated,
}

var ErrorTypeNames = map[ErrorType]string{
	Generic:            "Generic",
	Canceled:           "Canceled",
	InvalidArgument:    "InvalidArgument",
	DeadlineExceeded:   "DeadlineExceeded",
	NotFound:           "NotFound",
	AlreadyExists:      "AlreadyExists",
	PermissionDenied:   "PermissionDenied",
	ResourceExhausted:  "ResourceExhausted",
	FailedPrecondition: "FailedPrecondition",
	Aborted:            "Aborted",
	OutOfRange:         "OutOfRange",
	Unimplemented:      "Unimplemented",
	Internal:           "Internal",
	Unavailable:        "Unavailable",
	Unauthenticated:    "Unauthenticated",
}

// ValidErrorType checks the type enum to determine if the string belongs to the enumeration.
//   params:
//     errorType The type to be checked
//   returns:
//     Whether it is contained in the enum.
func ValidErrorType(errorType ErrorType) bool {
	_, exists := ErrorTypeNames[errorType]
	return exists
}

// ErrorTypeAsString returns the string representation of the error.
func ErrorTypeAsString(errorType ErrorType) string {
	s, _ := ErrorTypeNames[errorType]
	return s
}
