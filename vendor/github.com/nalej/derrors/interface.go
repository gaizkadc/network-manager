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

// Definition of the error interface.

package derrors

// Error defines the interface for all Daisho-defined errors.
type Error interface {
	// Error returns the string representation of the error. Notice that this particular method is required
	// in order to be be compatible with the default golang error.
	Error() string
	// Type returns the ErrorType associated with the current DaishoError.
	Type() ErrorType
	// DebugReport returns a detailed error report including the stack information.
	DebugReport() string
	// StackTrace returns an array with the calling stack that created the error.
	StackTrace() []StackEntry
}
