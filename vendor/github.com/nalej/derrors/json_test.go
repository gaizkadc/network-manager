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

// Serialization tests

package derrors

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestFromJsonGenericError(t *testing.T) {
	cause := errors.New("error cause")
	msg := "Error message"
	toSend := NewGenericError(msg, cause)

	data, err := json.Marshal(toSend)
	assertEquals(t, nil, err, "expecting no error")
	retrieved, err := FromJSON(data)

	assertEquals(t, nil, err, "message should be deserialized")
	//assertEquals(t, GenericErrorType, retrieved.Type(), "type mismatch")
	assertEquals(t, toSend, retrieved, "structure should match")
}

func TestFromJsonEntityError(t *testing.T) {

	cause := errors.New("another cause")
	msg := "Other message"
	toSend := NewGenericError(msg, cause)

	data, err := json.Marshal(toSend)
	assertEquals(t, nil, err, "expecting no error")
	retrieved, err := FromJSON(data)

	assertEquals(t, nil, err, "message should be deserialized")
	assertEquals(t, toSend, retrieved, "structure should match")

}

func TestFromJsonConnectionError(t *testing.T) {

	URL := "http://url-that-fails.com"
	cause := errors.New("yet another cause")
	msg := "Yet another message"
	toSend := NewUnavailableError(msg, cause).WithParams(URL)

	data, err := json.Marshal(toSend)
	assertEquals(t, nil, err, "expecting no error")
	retrieved, err := FromJSON(data)

	assertEquals(t, nil, err, "message should be deserialized")
	assertEquals(t, toSend, retrieved, "structure should match")

}

func TestFromJsonOperationError(t *testing.T) {

	param1 := "param1"
	param2 := "param2"
	param3 := "param3"
	cause := errors.New("operation failed")
	msg := "operation failure"
	toSend := NewInternalError(msg, cause).WithParams(param1, param2, param3)

	data, err := json.Marshal(toSend)
	assertEquals(t, nil, err, "expecting no error")
	retrieved, err := FromJSON(data)

	assertEquals(t, nil, err, "message should be deserialized")
	assertEquals(t, toSend, retrieved, "structure should match")

}
