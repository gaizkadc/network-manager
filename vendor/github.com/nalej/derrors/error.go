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

// Definition of the generic error.
// Notice that stack traces are not serialized.

package derrors

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"runtime"
)

// StackEntry structure that contains information about an element in the calling stack.
type StackEntry struct {
	// FunctionName of the calling function.
	FunctionName string
	// File where the function is located.
	File string
	// Line of the file where the function is located.
	Line int
}

// NewStackEntry creates a new stack entry.
//   params:
//     functionName The name of the calling function.
//     file The name of the file where the function is located.
//     line The line in the file.
func NewStackEntry(functionName string, file string, line int) *StackEntry {
	return &StackEntry{functionName, file, line}
}

// String returns the string representation of an StackEntry.
func (se *StackEntry) String() string {
	return fmt.Sprintf("%s - %s:%d", se.FunctionName, se.File, se.Line)
}

// GenericError structure that defines the basic elements shared by all DaishoErrors.
type GenericError struct {
	// ErrorType from the enumeration.
	ErrorType ErrorType `json:"errorType"`
	// Message contains the error message.
	Message string `json:"message"`
	// Parameters associated with error.
	Parameters []string `json:"parameters"`
	// causes contains the list of causes of the error.
	Causes []string `json:"causes"`
	// Parent Daisho Error.
	Parent interface{} `json:"parent"`
	// stackTrace contains the calling stack trace.
	Stack []StackEntry `json:"stackTrace"`
}

// WithParams permits to track extra parameters in the operation error.
func (ge *GenericError) WithParams(params ...interface{}) *GenericError {
	for _, value := range params {
		ser, err := json.Marshal(value)
		if err != nil {
			panic(err)
		}
		ge.Parameters = append(ge.Parameters, string(ser))
	}
	return ge
}

// CausedBy permits to link the error with a parent error. In this way, we can express the fact that a component
// fails cause another component fails.
func (ge *GenericError) CausedBy(parent Error) *GenericError {
	ge.Parent = parent
	return ge
}

// StackTraceAsString returns the stack trace elements as a string array.
func (ge *GenericError) StackTraceAsString() []string {
	result := make([]string, 0)
	for _, entry := range ge.Stack {
		result = append(result, entry.String())
	}
	return result
}

// StackToString generates a string with a stack entry per line.
func (ge *GenericError) StackToString() string {
	var buffer bytes.Buffer
	buffer.WriteString("StackTrace:\n")
	for i, v := range ge.Stack {
		sep := fmt.Sprintf("ST%d: ", i)
		buffer.WriteString(sep + v.String() + "\n")
	}
	return buffer.String()
}

// ParentError returns the parent error of the current Error or a standard golang error if the parent cannot be unmarshalled.
func (ge *GenericError) ParentError() (Error, error) {
	ser, err := json.Marshal(ge.Parent)
	if err != nil {
		return nil, err
	}
	return FromJSON(ser)
}

func (ge *GenericError) paramsToString() string {
	if len(ge.Parameters) == 0 {
		return ""
	}
	var buffer bytes.Buffer
	buffer.WriteString("Parameters:\n")
	for i, v := range ge.Parameters {
		sep := fmt.Sprintf("P%d: ", i)
		buffer.WriteString(sep + PrettyPrintStruct(v) + "\n")
	}
	return buffer.String()
}

func (ge *GenericError) causesToString() string {
	if len(ge.Causes) == 0 {
		return ""
	}
	var buffer bytes.Buffer
	buffer.WriteString("Caused by:\n")
	for i, v := range ge.Causes {
		sep := fmt.Sprintf("C%d: ", i)
		buffer.WriteString(sep + PrettyPrintStruct(v) + "\n")
	}
	return buffer.String()
}

func (ge *GenericError) parentToString() string {
	if ge.Parent == nil {
		return ""
	}
	var buffer bytes.Buffer
	buffer.WriteString("Parent:\n")
	parent, err := ge.ParentError()
	if err == nil {
		buffer.WriteString(parent.DebugReport())
	} else {
		buffer.WriteString("Cannot deserialize parent error:" + err.Error() + "\n")
	}
	return buffer.String()
}

func (ge *GenericError) Error() string {
	return fmt.Sprintf("[%s] %s", ErrorTypeAsString(ge.ErrorType), ge.Message)
}

// Type returns the ErrorType associated with the current DaishoError.
func (ge *GenericError) Type() ErrorType {
	return ge.ErrorType
}

// DebugReport returns a detailed error report including the stack information.
func (ge *GenericError) DebugReport() string {
	return fmt.Sprintf("%s\n%s\n%s\n%s\n%s",
		ge.Error(), ge.paramsToString(), ge.causesToString(), ge.StackToString(), ge.parentToString())
}

// StackTrace returns an array with the calling stack that created the error.
func (ge *GenericError) StackTrace() []StackEntry {
	return ge.Stack
}

// AsError checks an error. If it is nil, it returns nil, if not, it will create an equivalent GenericError
func AsError(err error, msg string) Error {
	if err != nil {
		return NewGenericError(msg, err)
	}
	return nil
}

// AsErrorWithParams checks an error. If it is nil, it returns nil, if not, it will create an equivalent
// GenericError with a given set of parameters.
func AsErrorWithParams(err error, msg string, params ...interface{}) Error {
	if err != nil {
		return NewGenericError(msg, err).WithParams(params)
	}
	return nil
}

// GetStackTrace retrieves the calling stack and transform that information into an array of StackEntry.
func GetStackTrace() []StackEntry {
	var programCounters [32]uintptr
	// Skip the three first callers.
	callersToSkip := 3
	callerCount := runtime.Callers(callersToSkip, programCounters[:])
	stackTrace := make([]StackEntry, callerCount)
	for i := 0; i < callerCount; i++ {
		function := runtime.FuncForPC(programCounters[i])
		filePath, line := function.FileLine(programCounters[i])
		stackTrace[i] = *NewStackEntry(function.Name(), filePath, line)
	}
	return stackTrace
}

// ErrorsToString transform a list of errors into a list of strings.
func ErrorsToString(errors []error) []string {
	result := make([]string, len(errors))
	for i := 0; i < len(errors); i++ {
		result[i] = errors[i].Error()
	}
	return result
}

// PrettyPrintStruct aims to produce a pretty print of a giving structure by analyzing it. If the data structure
// provides a String method, that will method will be invoked instead of the common string representation.
func PrettyPrintStruct(data interface{}) string {
	v := reflect.ValueOf(data)
	method := v.MethodByName("String")
	if method.IsValid() && method.Type().NumIn() == 0 && method.Type().NumOut() == 1 &&
		method.Type().Out(0).Kind() == reflect.String {
		result := method.Call([]reflect.Value{})
		return result[0].String()
	}
	return fmt.Sprintf("%#v", data)
}

func NewError(errorType ErrorType, msg string, causes ...error) *GenericError {
	return &GenericError{
		errorType,
		msg,
		make([]string, 0),
		ErrorsToString(causes),
		nil,
		GetStackTrace()}
}

// NewGenericError returns a general purpose error.
func NewGenericError(msg string, causes ...error) *GenericError {
	return &GenericError{
		Generic,
		msg,
		make([]string, 0),
		ErrorsToString(causes),
		nil,
		GetStackTrace()}
}

// NewCanceledError returns an error associated with an operation that has been canceled.
func NewCanceledError(msg string, causes ...error) *GenericError {
	return &GenericError{
		Canceled,
		msg,
		make([]string, 0),
		ErrorsToString(causes),
		nil,
		GetStackTrace()}
}

// NewInvalidArgumentError returns an error that indicates the use of an invalid argument.
func NewInvalidArgumentError(msg string, causes ...error) *GenericError {
	return &GenericError{
		InvalidArgument,
		msg,
		make([]string, 0),
		ErrorsToString(causes),
		nil,
		GetStackTrace()}
}

// NewDeadlineExceededError returns an error that indicates the deadline for the completion of an operation expired.
func NewDeadlineExceededError(msg string, causes ...error) *GenericError {
	return &GenericError{
		DeadlineExceeded,
		msg,
		make([]string, 0),
		ErrorsToString(causes),
		nil,
		GetStackTrace()}
}

// NewNotFoundError returns an error that indicates that the requested entity did not exists.
func NewNotFoundError(msg string, causes ...error) *GenericError {
	return &GenericError{
		NotFound,
		msg,
		make([]string, 0),
		ErrorsToString(causes),
		nil,
		GetStackTrace()}
}

// NewAlreadyExistsError returns an error that indicates that the target entity already exists.
func NewAlreadyExistsError(msg string, causes ...error) *GenericError {
	return &GenericError{
		AlreadyExists,
		msg,
		make([]string, 0),
		ErrorsToString(causes),
		nil,
		GetStackTrace()}
}

// NewPermissionDeniedError returns an error that indicates that the client is not authorized.
func NewPermissionDeniedError(msg string, causes ...error) *GenericError {
	return &GenericError{
		PermissionDenied,
		msg,
		make([]string, 0),
		ErrorsToString(causes),
		nil,
		GetStackTrace()}
}

// NewResourceExhaustedError returns an error that indicates that a given resource has been exhausted.
func NewResourceExhaustedError(msg string, causes ...error) *GenericError {
	return &GenericError{
		ResourceExhausted,
		msg,
		make([]string, 0),
		ErrorsToString(causes),
		nil,
		GetStackTrace()}
}

// NewFailedPreconditionError returns an error that indicates that a given precondition for an operation failed.
func NewFailedPreconditionError(msg string, causes ...error) *GenericError {
	return &GenericError{
		FailedPrecondition,
		msg,
		make([]string, 0),
		ErrorsToString(causes),
		nil,
		GetStackTrace()}
}

// NewAbortedError returns an error that indicates that a given operation was aborted due to an internal issue.
func NewAbortedError(msg string, causes ...error) *GenericError {
	return &GenericError{
		Aborted,
		msg,
		make([]string, 0),
		ErrorsToString(causes),
		nil,
		GetStackTrace()}
}

// NewOutOfRangeError returns an error that indicates that a requested resource is out of the available range.
func NewOutOfRangeError(msg string, causes ...error) *GenericError {
	return &GenericError{
		OutOfRange,
		msg,
		make([]string, 0),
		ErrorsToString(causes),
		nil,
		GetStackTrace()}
}

// NewUnimplementedError returns an error that indicates that a requested operation is not implemented yet.
func NewUnimplementedError(msg string, causes ...error) *GenericError {
	return &GenericError{
		Unimplemented,
		msg,
		make([]string, 0),
		ErrorsToString(causes),
		nil,
		GetStackTrace()}
}

// NewInternalError returns an error that indicates that an internal error occurred.
func NewInternalError(msg string, causes ...error) *GenericError {
	return &GenericError{
		Internal,
		msg,
		make([]string, 0),
		ErrorsToString(causes),
		nil,
		GetStackTrace()}
}

// NewUnavailableError returns an error that indicates that a given service is not currently available.
func NewUnavailableError(msg string, causes ...error) *GenericError {
	return &GenericError{
		Unavailable,
		msg,
		make([]string, 0),
		ErrorsToString(causes),
		nil,
		GetStackTrace()}
}

// NewUnauthenticatedError returns an error that indicates that a given request is not authenticated.
func NewUnauthenticatedError(msg string, causes ...error) *GenericError {
	return &GenericError{
		Unauthenticated,
		msg,
		make([]string, 0),
		ErrorsToString(causes),
		nil,
		GetStackTrace()}
}

// FromJSON unmarshalls a byte array with the JSON representation into an Error of the correct type.
//   params:
//     data The byte array with the serialized JSON.
//   returns:
//     An Error if the data can be unmarshalled.
//     A Golang error if the unmarshal operation fails.
func FromJSON(data []byte) (Error, error) {
	genericError := &GenericError{}
	err := json.Unmarshal(data, &genericError)
	if err != nil {
		return nil, err
	}
	if ValidErrorType(genericError.ErrorType) {
		var result Error = genericError
		return result, nil
	}
	return nil, errors.New("unsupported error type in conversion")
}
