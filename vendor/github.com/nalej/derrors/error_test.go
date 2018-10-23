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

// Error tests

package derrors

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
)

func callerGetStackTrace() []StackEntry {
	return GetStackTrace()
}

func TestGetStackTrace(t *testing.T) {
	stackTrace := callerGetStackTrace()
	assertTrue(t, len(stackTrace) > 0, "expecting stack")
	parent := stackTrace[0]
	parentFunctionName := strings.Split(parent.FunctionName, ".")[2]
	assertEquals(t, "TestGetStackTrace",
		parentFunctionName, "Expecting parent function")
}

type testPrettyStruct struct {
	msg string
}

func (ss *testPrettyStruct) String() string {
	return "PRETTY " + ss.msg
}

func TestNewGenericError(t *testing.T) {
	var err Error = NewGenericError("msg")
	assertTrue(t, err != nil, "Expecting new error")
}

func TestNewCanceledError(t *testing.T) {
	var err Error = NewCanceledError("msg")
	assertTrue(t, err != nil, "Expecting new error")
}

func TestNewInvalidArgumentError(t *testing.T) {
	var err Error = NewInvalidArgumentError("msg")
	assertTrue(t, err != nil, "Expecting new error")
}

func TestNewDeadlineExceededError(t *testing.T) {
	var err Error = NewDeadlineExceededError("msg")
	assertTrue(t, err != nil, "Expecting new error")
}

func TestNewNotFoundError(t *testing.T) {
	var err Error = NewNotFoundError("msg")
	assertTrue(t, err != nil, "Expecting new error")
}

func TestNewAlreadyExistsError(t *testing.T) {
	var err Error = NewAlreadyExistsError("msg")
	assertTrue(t, err != nil, "Expecting new error")
}

func TestNewPermissionDeniedError(t *testing.T) {
	var err Error = NewPermissionDeniedError("msg")
	assertTrue(t, err != nil, "Expecting new error")
}

func TestNewResourceExhaustedError(t *testing.T) {
	var err Error = NewResourceExhaustedError("msg")
	assertTrue(t, err != nil, "Expecting new error")
}

func TestNewFailedPreconditionError(t *testing.T) {
	var err Error = NewFailedPreconditionError("msg")
	assertTrue(t, err != nil, "Expecting new error")
}

func TestNewAbortedError(t *testing.T) {
	var err Error = NewAbortedError("msg")
	assertTrue(t, err != nil, "Expecting new error")
}

func TestNewOutOfRangeError(t *testing.T) {
	var err Error = NewOutOfRangeError("msg")
	assertTrue(t, err != nil, "Expecting new error")
}

func TestNewUnimplementedError(t *testing.T) {
	var err Error = NewUnimplementedError("msg")
	assertTrue(t, err != nil, "Expecting new error")
}

func TestNewInternalError(t *testing.T) {
	var err Error = NewInternalError("msg")
	assertTrue(t, err != nil, "Expecting new error")
}

func TestNewUnavailableError(t *testing.T) {
	var err Error = NewUnavailableError("msg")
	assertTrue(t, err != nil, "Expecting new error")
}

func TestNewUnauthenticatedError(t *testing.T) {
	var err Error = NewUnauthenticatedError("msg")
	assertTrue(t, err != nil, "Expecting new error")
}

func TestPrettyPrintStruct(t *testing.T) {
	basicElement := "string"
	r1 := PrettyPrintStruct(basicElement)
	assertEquals(t, "\"string\"", r1, "expecting same message")
	type basicStruct struct {
		msg string
	}
	structElement := &basicStruct{"string2"}
	r2 := PrettyPrintStruct(structElement)
	assertEquals(t, "&derrors.basicStruct{msg:\"string2\"}", r2, "expecting struct message")

	stringElement := &testPrettyStruct{"PRINT"}
	r3 := PrettyPrintStruct(stringElement)
	assertEquals(t, "PRETTY PRINT", r3, "expecting pretty print")

}

func TestError(t *testing.T) {
	parent := errors.New("golang error")
	parent2 := errors.New("previous error")
	e := NewGenericError("I/O error", parent, parent2)
	errorMsg := e.Error()
	assertEquals(t, "[Generic] I/O error", errorMsg, "Message should match")
	detailedError := e.DebugReport()
	fmt.Println(detailedError)
	fmt.Println("String(): " + e.Error())
}

func TestAsError(t *testing.T) {
	err := errors.New("some golang error")
	derror := AsError(err, "msg")
	assertEquals(t, "msg", derror.(*GenericError).Message, "Expecting message")

	derrorFromNil := AsError(nil, "msg")
	assertTrue(t, derrorFromNil == nil, "Should be nil")

	derrorWithParam := AsErrorWithParams(err, "msg", "param1")
	assertTrue(t, derrorWithParam != nil, "should not be nil")
	assertEquals(t, 1, len(derrorWithParam.(*GenericError).Parameters), "expecting one parameter")
}

func TestCausedBy(t *testing.T) {
	parent := NewGenericError("parent operation")
	e := NewGenericError("current operation").CausedBy(parent)
	assertTrue(t, e != nil, "Should not be nil")
	assertTrue(t, e.Parent != nil, "Expecting parent")
}

// AsError returned * GenericError instead of Error creating a problem when evaluating
// if the result is an error. assert.Nil reported the result as nil while err != nil failed.
func TestDP1092(t *testing.T) {
	//err need to be a DaishoError instead of a GenericError
	var err = AsError(nil, "Testing conversion")
	if err != nil {
		t.Error("Should be nil")
	}
}

type MockStruct struct {
	A string
	B bool
	C int
}

func NewMockStruct() *MockStruct {
	return &MockStruct{A: "aa", B: true, C: 10}
}

func TestDP1283(t *testing.T) {
	data := NewMockStruct()
	msg := "error valid parent"
	sysError := NewGenericError(msg).WithParams(NewMockStruct()).
		WithParams(NewGenericError("prb")).CausedBy(NewGenericError("Super parent"))

	err := NewGenericError("error valid").WithParams("id1").WithParams(data).CausedBy(sysError)

	debugReport := err.DebugReport()

	result, errSer := json.Marshal(err)

	if errSer != nil {
		t.Error("serialization must work")
	}

	errRecover := &GenericError{}
	errDes := json.Unmarshal(result, errRecover)
	if errDes != nil {
		t.Error("deserialization must work")
	}
	parent, errParent := errRecover.ParentError()
	if errParent != nil {
		t.Error("recover parent error must work")
	}
	if parent.Error() != sysError.Error() {
		t.Error("error must be equal")
	}
	debugReportRecover := errRecover.DebugReport()
	println(debugReportRecover)
	println(debugReport)
	if debugReport != debugReportRecover {
		t.Error("debug report must be equals")
	}

}

func TestDP1283UsingFromJson(t *testing.T) {
	data := NewMockStruct()
	sysError := NewGenericError("error valid")

	err := NewGenericError("error valid").WithParams("id1").WithParams(data).CausedBy(sysError)

	result, errSer := json.Marshal(err)

	if errSer != nil {
		t.Error("Serialization must work")
	}

	errRecover, errDes := FromJSON(result)
	if errDes != nil {
		t.Error("Deserialization must work")
	}
	if errRecover == nil {
		t.Error("Recover err is not empty")
	}

}

func TestDP1283WithoutObject(t *testing.T) {
	sysError := NewGenericError("error valid")

	err := NewGenericError("error valid").WithParams("id1").CausedBy(sysError)

	result, errSer := json.Marshal(err)

	if errSer != nil {
		t.Error("Serialization must work")
	}

	errRecover := &GenericError{}
	errDes := json.Unmarshal(result, errRecover)

	if errDes != nil {
		t.Error("Deserialization must work")
	}

}

func TestDP1283WithoutParams(t *testing.T) {
	sysError := NewGenericError("error valid")

	err := NewGenericError("error valid").CausedBy(sysError)

	result, errSer := json.Marshal(err)

	if errSer != nil {
		t.Error("Serialization must work")
	}

	errRecover := &GenericError{}
	errDes := json.Unmarshal(result, errRecover)
	if errDes != nil {
		t.Error("Deserialization must work")
	}

}

func TestDP1283WithoutAnything(t *testing.T) {
	err := NewGenericError("error valid")

	result, errSer := json.Marshal(err)

	if errSer != nil {
		t.Error("Serialization must work")
	}

	errRecover := &GenericError{}
	errDes := json.Unmarshal(result, errRecover)
	if errDes != nil {
		t.Error("Deserialization must work")
	}

}
