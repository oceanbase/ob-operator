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

package helper

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

type OceanBaseVersion struct {
	Major    int    `json:"major"`
	Minor    int    `json:"minor"`
	Patch    int    `json:"patch"`
	SubPatch int    `json:"subPatch"`
	Build    string `json:"build"`
}

func GetCurrentVersion(oceanbaseInstallPath string) (string, error) {
	output, err := exec.Command("bash", "-c", fmt.Sprintf("export LD_LIBRARY_PATH=%s/lib; %s/bin/observer -V", oceanbaseInstallPath, oceanbaseInstallPath)).CombinedOutput()
	if err != nil {
		return "", errors.Wrap(err, "Failed to execute version command")
	}
	fmt.Println(string(output))
	lines := strings.Split(string(output), "\n")
	if len(lines) > 3 {
		versionStr := strings.Split(lines[1], " ")
		version := versionStr[len(versionStr)-1]
		releaseStr := strings.Split(strings.Split(lines[3], " ")[1], "-")[0]
		return fmt.Sprintf("%s-%s", version[0:len(version)-1], releaseStr), nil
	} else {
		return "", errors.New("OB Version Format is Wrong")
	}
}

func ParseOceanBaseVersion(version string) (*OceanBaseVersion, error) {
	ver := &OceanBaseVersion{}
	var err error
	pattern := "^[0-9]+\\.[0-9]+\\.[0-9]+(\\.[0-9]+)?(-[0-9]+)?$"
	regex := regexp.MustCompile(pattern)
	if !regex.MatchString(version) {
		return nil, errors.New("version format is wrong, should be like " + pattern + " but got " + version)
	}

	verStr := strings.Split(version, "-")
	if len(verStr) > 1 {
		ver.Build = verStr[1]
	}

	verStr = strings.Split(verStr[0], ".")
	if len(verStr) < 3 {
		return nil, errors.New("version format is wrong")
	}

	ver.Major, err = strconv.Atoi(verStr[0])
	if err != nil {
		return nil, err
	}
	ver.Minor, err = strconv.Atoi(verStr[1])
	if err != nil {
		return nil, err
	}
	ver.Patch, err = strconv.Atoi(verStr[2])
	if err != nil {
		return nil, err
	}

	if len(verStr) > 3 {
		ver.SubPatch, err = strconv.Atoi(verStr[3])
		if err != nil {
			return nil, err
		}
	}
	return ver, nil
}

func (v *OceanBaseVersion) String() string {
	verPart := strings.TrimRight(fmt.Sprintf("%d.%d.%d.%d", v.Major, v.Minor, v.Patch, v.SubPatch), ".")
	if v.Build != "" {
		return fmt.Sprintf("%s-%s", verPart, v.Build)
	}
	return verPart
}

func (v *OceanBaseVersion) Cmp(other *OceanBaseVersion) int {
	if v.Major != other.Major {
		return v.Major - other.Major
	}
	if v.Minor != other.Minor {
		return v.Minor - other.Minor
	}
	if v.Patch != other.Patch {
		return v.Patch - other.Patch
	}
	if v.SubPatch != other.SubPatch {
		return v.SubPatch - other.SubPatch
	}
	if v.Build != other.Build {
		return strings.Compare(v.Build, other.Build)
	}
	return 0
}
