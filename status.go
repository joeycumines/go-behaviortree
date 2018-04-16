/*
   Copyright 2018 Joseph Cumines

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
 */

package behaviortree

import "fmt"

const (
	_              = iota
	Running Status = iota
	Success
	Failure
)

type (
	// Status is a type with three valid values, Running, Success, and Failure, the three possible states for BTs
	Status int
)

// Status returns the status value, but defaults to failure on out of bounds
func (s Status) Status() Status {
	switch s {
	case Running:
		return Running
	case Success:
		return Success
	default:
		return Failure
	}
}

// String returns a string representation of the status
func (s Status) String() string {
	switch s {
	case Running:
		return `running`
	case Success:
		return `success`
	case Failure:
		return `failure`
	default:
		return fmt.Sprintf("unknown status (%d)", s)
	}
}
