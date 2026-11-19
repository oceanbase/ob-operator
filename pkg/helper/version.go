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
	"path/filepath"
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

const numericVersion = `[0-9]+\.[0-9]+\.[0-9]+(?:\.[0-9]+)?`

var (
	normalizedVersion = regexp.MustCompile(`^` + numericVersion + `(?:-[0-9]+)?$`)
	observerBanner    = regexp.MustCompile(`(?m)^[ \t]*observer[ \t]+\([^()\r\n]+[ \t]+(` + numericVersion + `)\)[ \t\r]*$`)
	observerRevision  = regexp.MustCompile(`(?m)^[ \t]*REVISION:[ \t]*([0-9]+)(?:-[^\s]+)?[ \t\r]*$`)
)

func GetCurrentVersion(oceanbaseInstallPath string) (string, error) {
	cmd := exec.Command(filepath.Join(oceanbaseInstallPath, "bin", "observer"), "-V")
	cmd.Env = append(cmd.Environ(), "LD_LIBRARY_PATH="+filepath.Join(oceanbaseInstallPath, "lib"))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", errors.Wrap(err, "execute observer -V")
	}
	return ParseVersionOutput(string(output))
}

// ParseVersionOutput accepts the helper's single-line version and the full
// observer -V banner returned by older images.
func ParseVersionOutput(output string) (string, error) {
	version := observerBanner.FindStringSubmatch(output)
	revision := observerRevision.FindStringSubmatch(output)
	var reported string
	if version != nil && revision != nil {
		reported = version[1] + "-" + revision[1]
	}
	// Startup logging and the helper's version share the Kubernetes Job log.
	for line := range strings.SplitSeq(output, "\n") {
		line = strings.TrimSpace(line)
		if !normalizedVersion.MatchString(line) {
			continue
		}
		if reported != "" && reported != line {
			return "", errors.New("conflicting OceanBase versions in output")
		}
		reported = line
	}
	if reported == "" {
		return "", errors.New("observer output is missing a valid version or REVISION")
	}
	return reported, nil
}

func ParseOceanBaseVersion(version string) (*OceanBaseVersion, error) {
	ver := &OceanBaseVersion{}
	var err error
	if !normalizedVersion.MatchString(version) {
		return nil, errors.New("version format is wrong, should be like " + normalizedVersion.String() + " but got " + version)
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
