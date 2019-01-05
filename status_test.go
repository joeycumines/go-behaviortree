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

import (
	"fmt"
	"testing"
)

func TestStatus_String(t *testing.T) {
	testCases := []struct {
		Status Status
		String string
	}{
		{
			Status: Success,
			String: `success`,
		},
		{
			Status: Failure,
			String: `failure`,
		},
		{
			Status: Running,
			String: `running`,
		},
		{
			Status: 0,
			String: `unknown status (0)`,
		},
		{
			Status: 234,
			String: `unknown status (234)`,
		},
	}

	for i, testCase := range testCases {
		name := fmt.Sprintf("TestStatus_String_#%d", i)

		if actual := testCase.Status.String(); actual != testCase.String {
			t.Errorf("%s failed: expected stringer '%s' != actual '%s'", name, testCase.String, actual)
		}
	}
}

func TestStatus_Status(t *testing.T) {
	testCases := []struct {
		Status   Status
		Expected Status
	}{
		{
			Status:   Success,
			Expected: Success,
		},
		{
			Status:   Failure,
			Expected: Failure,
		},
		{
			Status:   Running,
			Expected: Running,
		},
		{
			Status:   0,
			Expected: Failure,
		},
		{
			Status:   234,
			Expected: Failure,
		},
	}

	for i, testCase := range testCases {
		name := fmt.Sprintf("TestStatus_Status_#%d", i)

		if actual := testCase.Status.Status(); actual != testCase.Expected {
			t.Errorf("%s failed: expected behaviortree.Status.Status '%s' != actual '%s'", name, testCase.Status, actual)
		}
	}
}
