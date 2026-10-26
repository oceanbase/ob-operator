package lakehouse

import (
	"flag"
	"fmt"
	"math"
	"math/big"
	"regexp"
	"strconv"

	"k8s.io/apimachinery/pkg/api/resource"
)

// Only explicitly supplied flags are recorded. An unset CLI option must not
// replace a quickstart default, values file, or saved installation setting.
type resourceOverrides map[string]string

type resourceOption struct {
	name, help string
	path       []string
	cpu        bool
}

var resourceOptions = []resourceOption{
	{"observer-cpu", "CPU per Observer, e.g. 2, 1.5 or 500m; overrides -f", []string{"observer", "resource", "cpu"}, true},
	{"observer-memory", "Memory per Observer, e.g. 16Gi; overrides -f", []string{"observer", "resource", "memory"}, false},
	{"observer-data-size", "Data volume per Observer, e.g. 100Gi; overrides -f", []string{"observer", "dataSize"}, false},
	{"observer-log-size", "Log volume per Observer, e.g. 20Gi; overrides -f", []string{"observer", "logSize"}, false},
	{"logservice-cpu", "CPU per LogService instance (three by default), e.g. 1; overrides -f", []string{"logService", "resource", "cpu"}, true},
	{"logservice-memory", "Memory per LogService instance (three by default), e.g. 4Gi; overrides -f", []string{"logService", "resource", "memory"}, false},
	{"logservice-store-size", "Store volume per LogService instance, e.g. 20Gi; overrides -f", []string{"logService", "storeSize"}, false},
	{"logservice-log-size", "Log volume per LogService instance, e.g. 10Gi; overrides -f", []string{"logService", "logSize"}, false},
	{"minio-size", "Bundled MinIO volume, e.g. 100Gi; overrides -f", []string{"minio", "size"}, false},
}

var (
	resourceCPUFormat  = regexp.MustCompile(`^([0-9]+(\.[0-9]{1,3})?|[0-9]+m)$`)
	resourceSizeFormat = regexp.MustCompile(`^([1-9][0-9]*)(Mi|Gi|Ti)$`)
)

func (overrides *resourceOverrides) register(fs *flag.FlagSet) {
	*overrides = resourceOverrides{}
	for _, option := range resourceOptions {
		option := option
		fs.Func(option.name, option.help, func(value string) error {
			normalized, err := normalizeResourceOption(option, value)
			if err != nil {
				return err
			}
			(*overrides)[option.name] = normalized
			return nil
		})
	}
}

func normalizeResourceOption(option resourceOption, value string) (string, error) {
	if option.cpu {
		if !resourceCPUFormat.MatchString(value) {
			return "", fmt.Errorf("--%s requires positive CPU cores or millicores, e.g. 2, 1.5 or 500m (minimum 1m)", option.name)
		}
		q, err := resource.ParseQuantity(value)
		if err != nil || q.Sign() <= 0 || q.Cmp(resource.MustParse("1m")) < 0 || q.Cmp(resource.MustParse("1000000000")) > 0 {
			return "", fmt.Errorf("--%s must be between 1m and 1000000000 CPU cores", option.name)
		}
		// The chart accepts whole cores and millicores, not Kubernetes' other
		// canonical suffixes (for example 1k or 1G).
		milli := q.MilliValue() // Bounded above, so this cannot overflow.
		if milli%1000 == 0 {
			return strconv.FormatInt(milli/1000, 10), nil
		}
		return strconv.FormatInt(milli, 10) + "m", nil
	}
	match := resourceSizeFormat.FindStringSubmatch(value)
	if match == nil {
		return "", fmt.Errorf("--%s requires a positive whole-number size with Mi, Gi or Ti, e.g. 16Gi", option.name)
	}
	// ParseQuantity clamps oversized binary quantities. Check the original
	// number first so overflow cannot silently become a different requested size.
	bytes, ok := new(big.Int).SetString(match[1], 10)
	if !ok {
		return "", fmt.Errorf("invalid --%s size", option.name)
	}
	bytes.Lsh(bytes, map[string]uint{"Mi": 20, "Gi": 30, "Ti": 40}[match[2]])
	if bytes.Cmp(big.NewInt(math.MaxInt64)) > 0 {
		return "", fmt.Errorf("--%s exceeds the supported size range", option.name)
	}
	return value, nil
}

func (overrides resourceOverrides) apply(values map[string]interface{}) map[string]interface{} {
	result := mergeValues(values, nil)
	for _, option := range resourceOptions {
		if value, supplied := overrides[option.name]; supplied {
			setValue(result, value, option.path...)
		}
	}
	return result
}

// Equivalent quantities can have different textual forms. In particular MinIO
// size participates in the chart's storage-identity hash, so replays must retain
// the installed representation, not merely an equal number of bytes.
func (overrides resourceOverrides) reuseInstalled(values, installed map[string]interface{}) map[string]interface{} {
	result := mergeValues(values, nil)
	for _, option := range resourceOptions {
		if _, supplied := overrides[option.name]; supplied {
			setValue(result, pfGet(installed, option.path...), option.path...)
		}
	}
	return result
}

func (overrides resourceOverrides) validateContext(values, installed map[string]interface{}, existing bool) error {
	if _, supplied := overrides["minio-size"]; supplied {
		if pfString(pfGet(values, "objectStorage", "mode")) != "bundled" {
			return fmt.Errorf("--minio-size applies only to bundled object storage; external storage capacity is managed separately")
		}
		if pfString(pfGet(values, "minio", "existingClaim")) != "" {
			return fmt.Errorf("--minio-size cannot resize an existingClaim; manage that PVC separately")
		}
	}
	if !existing {
		return nil
	}
	for _, option := range resourceOptions {
		value, supplied := overrides[option.name]
		if !supplied {
			continue
		}
		previous, err := resource.ParseQuantity(pfString(pfGet(installed, option.path...)))
		next := resource.MustParse(value) // Validated when parsing the CLI.
		if err != nil || previous.Cmp(next) != 0 {
			return fmt.Errorf("--%s changes the sizing of an existing release; sizing flags configure a new installation or repeat its current settings, not resize existing resources", option.name)
		}
	}
	return nil
}
