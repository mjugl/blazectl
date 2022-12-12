// Copyright 2019 - 2022 The Samply Community
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package util

import (
	"fmt"
	"math"
	"sort"
	"time"

	"gonum.org/v1/gonum/floats"
)

// DurationStatistics represents statistics about measured durations.
// Comprises information about the mean and max as well as different
// percentiles (50, 95 and 99).
type DurationStatistics struct {
	Mean, Q50, Q95, Q99, Max time.Duration
}

// Calculates the DurationStatistics for a set of given durations.
func CalculateDurationStatistics(durations []float64) DurationStatistics {
	var filteredDurations []float64

	for _, d := range durations {
		if !math.IsNaN(d) && d >= 0 {
			filteredDurations = append(filteredDurations, d)
		}
	}

	if len(filteredDurations) == 0 {
		return DurationStatistics{}
	}

	sort.Float64s(filteredDurations)
	return DurationStatistics{
		Mean: time.Duration(floats.Sum(filteredDurations)/float64(len(filteredDurations))*1000) * time.Millisecond,
		Q50:  time.Duration(filteredDurations[len(filteredDurations)/2]*1000) * time.Millisecond,
		Q95:  time.Duration(filteredDurations[int(float32(len(filteredDurations))*0.95)]*1000) * time.Millisecond,
		Q99:  time.Duration(filteredDurations[int(float32(len(filteredDurations))*0.99)]*1000) * time.Millisecond,
		Max:  time.Duration(filteredDurations[len(filteredDurations)-1]*1000) * time.Millisecond,
	}
}

// FmtBytesHumanReadable takes an amount of bytes and returns them in a human readable form
// up to a unit of PiB.
func FmtBytesHumanReadable(bytes float32) string {
	units := []string{"B", "KiB", "MiB", "GiB", "TiB", "PiB"}

	var unitIdx int
	for {
		if bytes <= 1024 || (unitIdx+1) > len(units)-1 {
			break
		}

		bytes = bytes / 1024
		unitIdx++
	}

	return fmt.Sprintf("%.2f %s", bytes, units[unitIdx])
}

// FmtDurationHumanReadable takes a duration and returns it in a human readable form.
// This is basically equivalent to time.Duration.Round(time.Second) with the following differences:
//   - durations under a minute get printed with millisecond precision
//   - durations equal or above a minute get printed with second precision
func FmtDurationHumanReadable(d time.Duration) string {
	if d.Milliseconds() < 60000 {
		return fmt.Sprintf("%s", d.Round(time.Millisecond))
	} else {
		return fmt.Sprintf("%s", d.Round(time.Second))
	}
}
