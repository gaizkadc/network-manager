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

// NewGenericError returns a general purpose error.
func NewGenericError(msg string, causes ...error) *GenericError {
	return &GenericError{
		GenericErrorType,
		msg,
		make([]string, 0),
		ErrorsToString(causes),
		nil,
		GetStackTrace()}
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
	return fmt.Sprintf("[%s] %s", ge.ErrorType, ge.Message)
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

// NewEntityError creates a new DaishoError of type Entity.
//   params:
//     entity The associated entity.
//     msg The error message.
//   returns:
//     An EntityError.
func NewEntityError(entity interface{}, msg string, causes ...error) *GenericError {
	params := make([]string, 0)
	err := &GenericError{
		EntityErrorType,
		msg,
		params,
		ErrorsToString(causes),
		nil,
		GetStackTrace()}

	return err.WithParams(entity)
}

// NewConnectionError creates a new DaishoError of type Connection.
//   params:
//     msg The error message.
//   returns:
//     An EntityError.
func NewConnectionError(msg string, causes ...error) *GenericError {
	return &GenericError{
		ConnectionErrorType,
		msg,
		make([]string, 0),
		ErrorsToString(causes),
		nil,
		GetStackTrace()}
}

// NewOperationError creates a new OperationError.
func NewOperationError(msg string, causes ...error) *GenericError {
	return &GenericError{
		OperationErrorType,
		msg,
		make([]string, 0),
		ErrorsToString(causes),
		nil,
		GetStackTrace()}
}

// FromJSON unmarshalls a byte array with the JSON representation into a DaishoError of the correct type.
//   params:
//     data The byte array with the serialized JSON.
//   returns:
//     A DaishoError if the data can be unmarshalled.
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

// IsDecoded checks that the required fields are decoded.
func IsDecoded(err Error) bool {
	return err.Type() != ""
}
