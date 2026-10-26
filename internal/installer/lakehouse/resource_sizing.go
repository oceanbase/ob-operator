package lakehouse

import (
	"fmt"
	"math/big"
	"regexp"
	"strings"

	"k8s.io/apimachinery/pkg/api/resource"

	oceanbaseconst "github.com/oceanbase/ob-operator/internal/const/oceanbase"
)

var quickstartSizePattern = regexp.MustCompile(`^([1-9][0-9]*)(Mi|Gi|Ti)$`)

// validateQuickstartSizing checks the bundled Operator defaults for a new CLI
// sizing request. It does not discover a separately configured Operator's limits
// or change any parameter, resource request, or disk size supplied by the user.
func validateQuickstartSizing(values map[string]interface{}) error {
	observerMemory, err := quickstartSize(values, "observer", "resource", "memory")
	if err != nil {
		return err
	}
	observerData, err := quickstartSize(values, "observer", "dataSize")
	if err != nil {
		return err
	}
	observerLog, err := quickstartSize(values, "observer", "logSize")
	if err != nil {
		return err
	}
	for _, minimum := range []struct {
		field string
		value resource.Quantity
		min   string
	}{
		{"observer.resource.memory", observerMemory, oceanbaseconst.MinMemorySizeS},
		{"observer.dataSize", observerData, oceanbaseconst.MinDataDiskSizeS},
		{"observer.logSize", observerLog, oceanbaseconst.MinLogDiskSizeS},
	} {
		if minimum.value.Cmp(resource.MustParse(minimum.min)) < 0 {
			return fmt.Errorf("quickstart sizing: %s must be at least %s under the bundled Operator defaults", minimum.field, minimum.min)
		}
	}

	observerLimit, err := quickstartMemoryLimit(values, "observer", observerMemory, true)
	if err != nil {
		return err
	}
	requiredData := observerLimit.DeepCopy()
	requiredData.Add(observerLimit)
	requiredData.Add(observerLimit)
	if observerData.Cmp(requiredData) < 0 {
		return fmt.Errorf("quickstart sizing: observer.dataSize must be at least three times its effective memory_limit (%s; required data size %s); increase --observer-data-size or reduce --observer-memory / observer.parameters memory_limit; the bundled Operator defaults are used when memory_limit is not set", observerLimit.String(), requiredData.String())
	}

	logServiceMemory, err := quickstartSize(values, "logService", "resource", "memory")
	if err != nil {
		return err
	}
	// LogService has no Observer-style minimum or three-times disk constraint.
	_, err = quickstartMemoryLimit(values, "logService", logServiceMemory, false)
	return err
}

func quickstartSize(values map[string]interface{}, path ...string) (resource.Quantity, error) {
	field := strings.Join(path, ".")
	raw, ok := pfGet(values, path...).(string)
	parts := quickstartSizePattern.FindStringSubmatch(raw)
	if !ok || len(parts) != 3 {
		return resource.Quantity{}, fmt.Errorf("quickstart sizing: %s must be a positive integer size with Mi, Gi, or Ti units in complete values", field)
	}
	amount, ok := new(big.Int).SetString(parts[1], 10)
	if !ok {
		return resource.Quantity{}, fmt.Errorf("quickstart sizing: %s is not a valid size", field)
	}
	shift := map[string]uint{"Mi": 20, "Gi": 30, "Ti": 40}[parts[2]]
	amount.Lsh(amount, shift)
	if !amount.IsInt64() {
		return resource.Quantity{}, fmt.Errorf("quickstart sizing: %s exceeds the supported int64 byte capacity", field)
	}
	return *resource.NewQuantity(amount.Int64(), resource.BinarySI), nil
}

func quickstartMemoryLimit(values map[string]interface{}, component string, memory resource.Quantity, defaultLimit bool) (resource.Quantity, error) {
	var limitValue interface{}
	explicit := false
	if raw := pfGet(values, component, "parameters"); raw != nil {
		parameters, ok := raw.([]interface{})
		if !ok {
			return resource.Quantity{}, fmt.Errorf("quickstart sizing: %s.parameters must be a parameter list in complete values", component)
		}
		for i, rawParameter := range parameters {
			parameter, ok := rawParameter.(map[string]interface{})
			if !ok {
				return resource.Quantity{}, fmt.Errorf("quickstart sizing: %s.parameters[%d] must be a parameter object", component, i)
			}
			name, ok := parameter["name"].(string)
			if !ok || name == "" {
				return resource.Quantity{}, fmt.Errorf("quickstart sizing: %s.parameters[%d].name must be a non-empty string", component, i)
			}
			if name == "memory_limit" {
				// The Operator overlays parameters in list order: the last wins.
				limitValue, explicit = parameter["value"], true
			}
		}
	}
	if explicit {
		raw, ok := limitValue.(string)
		limit, err := resource.ParseQuantity(raw)
		if !ok || err != nil || limit.Sign() <= 0 {
			return resource.Quantity{}, fmt.Errorf("quickstart sizing: %s.parameters memory_limit must be a positive Kubernetes quantity string", component)
		}
		if limit.Cmp(memory) > 0 {
			return resource.Quantity{}, fmt.Errorf("quickstart sizing: %s.parameters memory_limit exceeds %s.resource.memory; increase the component memory or lower memory_limit", component, component)
		}
		return limit, nil
	}
	if !defaultLimit {
		return resource.Quantity{}, nil
	}
	// Match the bundled default exactly: floor(bytes * 90 / 100 / 2^20)
	// followed by a decimal M suffix. Big integers prevent intermediate overflow.
	amount := new(big.Int).SetInt64(memory.Value())
	amount.Mul(amount, big.NewInt(oceanbaseconst.DefaultMemoryLimitPercent))
	amount.Quo(amount, big.NewInt(100*oceanbaseconst.MegaConverter))
	limit, err := resource.ParseQuantity(amount.String() + "M")
	if err != nil || limit.Sign() <= 0 {
		return resource.Quantity{}, fmt.Errorf("quickstart sizing: cannot derive observer memory_limit from the bundled Operator defaults")
	}
	return limit, nil
}
