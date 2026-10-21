package lakehouse

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"sigs.k8s.io/yaml"
)

func readValues(path string) (map[string]interface{}, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var values map[string]interface{}
	if err = yaml.UnmarshalStrict(data, &values); err != nil {
		return nil, fmt.Errorf("invalid values file %s (check YAML types, indentation and duplicate keys)", path)
	}
	if values == nil {
		values = map[string]interface{}{}
	}
	return values, nil
}

// mergeValues follows Helm's map merge and list replacement semantics. Nulls
// remove inherited values; the caller never modifies its input maps.
func mergeValues(base, overlay map[string]interface{}) map[string]interface{} {
	out := map[string]interface{}{}
	for key, value := range base {
		if m, ok := value.(map[string]interface{}); ok {
			out[key] = mergeValues(m, nil)
		} else {
			out[key] = value
		}
	}
	for key, value := range overlay {
		if m, ok := value.(map[string]interface{}); ok {
			old, _ := out[key].(map[string]interface{})
			out[key] = mergeValues(old, m)
		} else {
			out[key] = value
		}
	}
	return out
}

func setValue(v map[string]interface{}, value interface{}, path ...string) {
	for _, key := range path[:len(path)-1] {
		next, ok := v[key].(map[string]interface{})
		if !ok {
			next = map[string]interface{}{}
			v[key] = next
		}
		v = next
	}
	v[path[len(path)-1]] = value
}

func defaultValues(bundle string) (map[string]interface{}, error) {
	child, err := readValues(filepath.Join(bundle, "charts/oceanbase-dashboard/values.yaml"))
	if err != nil {
		return nil, err
	}
	parent, err := readValues(filepath.Join(bundle, "charts/oceanbase-lakehouse/values.yaml"))
	if err != nil {
		return nil, err
	}
	return mergeValues(map[string]interface{}{"dashboard": child}, parent), nil
}

func initialValues(bundle string) (map[string]interface{}, error) {
	values, err := readValues(filepath.Join(bundle, "values/quickstart.yaml"))
	if err != nil {
		return nil, err
	}
	// Persist these before the first Kubernetes mutation so interrupted installs
	// reuse their original identities. Do not regenerate them for existing releases.
	var raw [4]byte
	if _, err = rand.Read(raw[:]); err != nil {
		return nil, err
	}
	id := int64(binary.BigEndian.Uint32(raw[:])%900000000 + 100000000)
	setValue(values, id, "logService", "clusterId")
	setValue(values, id+1, "observer", "clusterId")
	return values, nil
}

func writePrivate(path string, data []byte) error {
	dir := filepath.Dir(path)
	if err := privateDirectory(dir); err != nil {
		return err
	}
	file, err := os.CreateTemp(dir, ".install-*")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	if _, err = file.Write(data); err != nil {
		file.Close()
		return err
	}
	if err = file.Sync(); err != nil {
		file.Close()
		return err
	}
	if err = file.Close(); err != nil {
		return err
	}
	return os.Rename(file.Name(), path)
}

// Preserve the effective database/storage configuration even when it originally
// came from a chart default. New integration defaults remain available for keys
// which did not exist in the installed release.
func retainConfiguration(effective, user map[string]interface{}) map[string]interface{} {
	kept := map[string]interface{}{}
	for _, key := range []string{"storageClass", "imagePullSecrets", "objectStorage", "minio", "initialization", "logService", "observer", "scheduling", "dashboard"} {
		if value, ok := effective[key]; ok {
			kept[key] = value
		}
	}
	return mergeValues(kept, user)
}

func validateConfiguredImages(bundle string, values map[string]interface{}) error {
	data, err := os.ReadFile(filepath.Join(bundle, "image-lock.json"))
	if err != nil {
		return err
	}
	var lock struct {
		Images []struct{ Role, Reference string } `json:"images"`
	}
	if err = json.Unmarshal(data, &lock); err != nil {
		return err
	}
	refs := map[string]string{}
	for _, item := range lock.Images {
		refs[item.Role] = item.Reference
	}
	paths := map[string][]string{"observer": {"observer", "image"}, "logservice": {"logService", "image"}}
	if enabled, _ := pfGet(values, "dashboard", "enabled").(bool); enabled {
		paths["dashboard"] = []string{"dashboard", "image"}
	}
	if pfGet(values, "objectStorage", "mode") == "bundled" {
		paths["minio"] = []string{"minio", "image"}
	}
	paths["mc"] = []string{"initialization", "image"}
	for role, path := range paths {
		if refs[role] == "" || refs[role] != pfString(pfGet(values, path...)) {
			return fmt.Errorf("%s image differs from this package's image lock; use matching values or the advanced chart workflow with your own verified images", role)
		}
	}
	return nil
}

func privateDirectory(dir string) error {
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	st, err := os.Lstat(dir)
	if err != nil {
		return err
	}
	if !st.IsDir() || st.Mode()&os.ModeSymlink != 0 {
		return errors.New("installer state directory must be a real directory")
	}
	return os.Chmod(dir, 0700)
}
