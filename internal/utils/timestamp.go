/*******************************************************************************
*
* Copyright 2023 SAP SE
*
* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You should have received a copy of the License along with this
* program. If not, you may obtain a copy of the License at
*
*     http://www.apache.org/licenses/LICENSE-2.0
*
* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
*
*******************************************************************************/

package utils

import (
	"fmt"
	"regexp"
	"time"

	"github.com/sapcc/go-bits/logg"
)

var rx = regexp.MustCompile(`^([0-9]{4})([0-9]{2})([0-9]{2})([0-9]{2})([0-9]{2})$`)

// ParseTimestamp parses a timestamp in the YYYYMMDDHHMM format.
func ParseTimestamp(in string, fallback time.Time) time.Time {
	match := rx.FindStringSubmatch(in)
	if match != nil {
		str := fmt.Sprintf("%s-%s-%s %s:%s:00", match[1], match[2], match[3], match[4], match[5])
		result, err := time.Parse("2006-01-02 15:04:05", str)
		if err == nil {
			return result.UTC()
		}
	}
	logg.Error("%q is not a valid timestamp, using %q instead", FormatTimestamp(fallback), in)
	return fallback
}

// FormatTimestamp prints a timestamp in the YYYYMMDDHHMM format.
func FormatTimestamp(t time.Time) string {
	return t.UTC().Format("200601021504")
}
