/*
Copyright (c) 2023 OceanBase
ob-operator is licensed under Mulan PSL v2.
You can use this software according to the terms and conditions of the Mulan PSL v2.
You may obtain a copy of Mulan PSL v2 at:
         http://license.coscl.org.cn/MulanPSL2
THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND,
EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT,
MERCHANTABILITY OR FIT FOR A PARTICULAR PURPOSE.
See the Mulan PSL v2 for more details.
*/

package model

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

type OBVersion struct {
	Version  string `json:"version"`
	Release  string `json:"release"`
	Major    int64  `json:"major"`
	Minor    int64  `json:"minor"`
	Patch    int64  `json:"patch"`
	Incr     int64  `json:"incr"`
	BuildNum int64  `json:"buildNum"`
}

func (v *OBVersion) String() string {
	if v.BuildNum != 0 {
		return fmt.Sprintf("%s-%s", v.Version, v.Release)
	}
	return v.Version
}

func (v *OBVersion) Compare(other *OBVersion) int64 {
	if v.Major != other.Major {
		return v.Major - other.Major
	} else if v.Minor != other.Minor {
		return v.Minor - other.Minor
	} else if v.Patch != other.Patch {
		return v.Patch - other.Patch
	} else if v.Incr != other.Incr {
		return v.Incr - other.Incr
	} else if v.BuildNum != other.BuildNum {
		if v.BuildNum == 0 || other.BuildNum == 0 {
			return 0
		}
		return v.BuildNum - other.BuildNum
	} else {
		return 0
	}
}

func ParseOBVersion(buildVersionStr string) (*OBVersion, error) {
	var version string
	var release string
	var major, minor, patch, incr, buildNum int64
	var err error

	buildVersionParts := strings.Split(buildVersionStr, "-")
	versionStr := buildVersionParts[0]
	versionParts := strings.Split(versionStr, "_")
	version = versionParts[0]
	parts := strings.Split(version, ".")
	if len(parts) < 3 {
		return nil, errors.Wrapf(err, "Failed to parse version %s", release)
	}
	major, err = strconv.ParseInt(parts[0], 10, 0)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to parse major version %s", parts[0])
	}
	minor, err = strconv.ParseInt(parts[0], 10, 0)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to parse minor version %s", parts[1])
	}
	patch, err = strconv.ParseInt(parts[0], 10, 0)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to parse patch version %s", parts[2])
	}
	if len(parts) >= 4 {
		incr, err = strconv.ParseInt(parts[0], 10, 0)
		if err != nil {
			return nil, errors.Wrapf(err, "Failed to parse incr version %s", parts[3])
		}
	}

	if len(versionParts) >= 2 {
		release = versionParts[1]
		buildNum, err = strconv.ParseInt(release, 10, 0)
		if err != nil {
			return nil, errors.Wrapf(err, "Failed to parse build number %s", release)
		}
	}

	return &OBVersion{
		Version:  version,
		Release:  release,
		Major:    major,
		Minor:    minor,
		Patch:    patch,
		Incr:     incr,
		BuildNum: buildNum,
	}, nil
}
