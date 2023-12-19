// Copyright The nextver Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package version

import (
	"fmt"
	"strconv"
	"strings"
)

type Semver struct {
	prefix bool
	major  string
	minor  string
	patch  string
	meta   string
	pre    string
}

func ParseVersion(v string, strict bool) (Semver, error) {
	//if len(v) == 0 {
	//	return nil, ErrEmptyString
	//}

	s := Semver{}

	if strings.HasPrefix(v, "v") {
		s.prefix = true
		v = v[1:]
	}

	// Split the parts into [0]major, [1]minor, and [2]patch,prerelease,build
	split := strings.SplitN(v, ".", 3)
	if len(split) != 3 {
		return s, fmt.Errorf("invalid semver: %s", v)
	}

	// check for prerelease or build metadata
	var extra []string
	if strings.ContainsAny(split[2], "-+") {
		extra = strings.SplitN(split[2], "+", 2)
		if len(extra) > 1 {
			// build metadata
			s.meta = extra[1]
			split[2] = extra[0]
		}

		extra = strings.SplitN(split[2], "-", 2)
		if len(extra) > 1 {
			// prerelease
			s.pre = extra[1]
			split[2] = extra[0]
		}
	}

	var err error
	s.major, err = validateVersionPart(split[0], strict)
	if err != nil {
		return s, err
	}
	s.minor, err = validateVersionPart(split[1], strict)
	if err != nil {
		return s, err
	}
	s.patch, err = validateVersionPart(split[2], strict)
	if err != nil {
		return s, err
	}
	return s, nil
}

func validateVersionPart(part string, strict bool) (string, error) {
	if _, err := strconv.ParseUint(part, 10, 64); err != nil {
		if strict {
			return "", err
		}
		if part != "x" && part != "?" {
			return "", fmt.Errorf("version part must be a non-negative number or a single 'x' or '?'")
		}
	}
	return part, nil
}

func (s Semver) String() string {
	sb := &strings.Builder{}
	if s.prefix {
		sb.WriteString("v")
	}
	sb.WriteString(s.major)
	sb.WriteString(".")
	sb.WriteString(s.minor)
	sb.WriteString(".")
	sb.WriteString(s.patch)

	if s.pre != "" {
		sb.WriteString("-")
		sb.WriteString(s.pre)
	}
	if s.meta != "" {
		sb.WriteString("+")
		sb.WriteString(s.meta)
	}
	return sb.String()
}

func (s Semver) ApplyPattern(p string) (Semver, error) {
	split := strings.Split(p, "x")
	if len(split) == 1 {
		return Semver{}, fmt.Errorf("no increment indicator ('x') found in pattern: %s", p)
	}
	if len(split) > 2 {
		return Semver{}, fmt.Errorf("more than 1 increment indicator ('x') found in pattern: %s", p)
	}

	patternVersion, err := ParseVersion(p, false)
	if err != nil {
		return Semver{}, err
	}

	major, err := computeSegmentValue(patternVersion.major, s.major, true)
	if err != nil {
		return Semver{}, err
	}
	minor, err := computeSegmentValue(patternVersion.minor, s.minor, s.major == major)
	if err != nil {
		return Semver{}, err
	}
	patch, err := computeSegmentValue(patternVersion.patch, s.patch, s.major == major && s.minor == minor)
	if err != nil {
		return Semver{}, err
	}

	result := Semver{
		prefix: patternVersion.prefix,
		major:  major,
		minor:  minor,
		patch:  patch,
	}
	if result.LessThan(s) {
		return Semver{}, errVersionDecrease
	}
	return result, nil
}

func computeSegmentValue(patternSegment string, versionSegment string, allowIncrement bool) (string, error) {
	switch patternSegment {
	case "x":
		if allowIncrement {
			i, err := strconv.ParseUint(versionSegment, 10, 64)
			if err != nil {
				return "", err
			}
			i++
			return strconv.FormatUint(i, 10), nil
		}
		return "0", nil
	case "?":
		return versionSegment, nil
	default:
		return patternSegment, nil
	}
}

// LessThan tests if one version is less than another one.
func (s Semver) LessThan(o Semver) bool {
	return s.Compare(o) < 0
}

// GreaterThan tests if one version is greater than another one.
func (s Semver) GreaterThan(o Semver) bool {
	return s.Compare(o) > 0
}

// Equal tests if two versions are equal to each other.
// Note, versions can be equal with different metadata since metadata
// is not considered part of the comparable version.
func (s Semver) Equal(o Semver) bool {
	return s.Compare(o) == 0
}

// Compare compares this version to another one. It returns -1, 0, or 1 if
// the version smaller, equal, or larger than the other version.
//
// Versions are compared by X.Y.Z. Build metadata is ignored. Prerelease is
// lower than the version without a prerelease. Compare always takes into account
// prereleases. If you want to work with ranges using typical range syntaxes that
// skip prereleases if the range is not looking for them use constraints.
func (s Semver) Compare(o Semver) int {
	// Compare the major, minor, and patch version for differences. If a
	// difference is found return the comparison.
	if d := mustCompareSegmentValue(s.major, o.major); d != 0 {
		return d
	}
	if d := mustCompareSegmentValue(s.minor, o.minor); d != 0 {
		return d
	}
	if d := mustCompareSegmentValue(s.patch, o.patch); d != 0 {
		return d
	}

	// At this point the major, minor, and patch versions are the same.
	ps := s.pre
	po := o.pre

	if ps == "" && po == "" {
		return 0
	}
	if ps == "" {
		return 1
	}
	if po == "" {
		return -1
	}

	//return comparePrerelease(ps, po)
	return 0
}

func mustCompareSegmentValue(v, o string) int {
	vI, err := strconv.ParseUint(v, 10, 64)
	if err != nil {
		panic(err)
	}
	oI, err := strconv.ParseUint(o, 10, 64)
	if err != nil {
		panic(err)
	}

	if vI < oI {
		return -1
	}
	if vI > oI {
		return 1
	}

	return 0
}
