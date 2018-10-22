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

// Definition of the supported types of error.

package derrors

// ErrorType defines a new type for creating a enum of error types.
type ErrorType string

// GenericErrorType to be used with general errors.
const GenericErrorType ErrorType = "GenericError"

// ConnectionErrorType is associated with connectivity errors.
const ConnectionErrorType ErrorType = "Connection"

// EntityErrorType is associated with entity related errors including validation, association, etc.
const EntityErrorType ErrorType = "Entity"

// OperationErrorType is associated with failures in external operations.
const OperationErrorType ErrorType = "Operation"

// ProviderErrorType is associated with provider related errors including invalid operations, provider failures, etc.
const ProviderErrorType ErrorType = "Provider"

// OrchestrationErrorType is associated with orchestration related errors including orchestration failures, preconditions, etc.
const OrchestrationErrorType ErrorType = "Orchestration"

// ValidErrorType checks the type enum to determine if the string belongs to the enumeration.
//   params:
//     errorType The type to be checked
//   returns:
//     Whether it is contained in the enum.
func ValidErrorType(errorType ErrorType) bool {
	switch errorType {
	case "":
		return false
	case GenericErrorType:
		return true
	case ConnectionErrorType:
		return true
	case EntityErrorType:
		return true
	case OperationErrorType:
		return true
	case ProviderErrorType:
		return true
	case OrchestrationErrorType:
		return true
	default:
		return false
	}
}
