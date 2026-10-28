package lakehouse

import (
	"fmt"
	"os"
	"path/filepath"
)

type kubeconfigContextKey struct{}

// An explicit selection always wins. Do not replace an environment-selected
// cluster or a user's default config with the local K3s cluster. This selection
// is process-local: no ~/.kube/config, permission or user environment is changed.
func installerKubeconfig(explicit, environment, k3s string) (string, error) {
	if explicit != "" {
		path, err := filepath.Abs(explicit)
		if err != nil {
			return "", err
		}
		file, err := os.Open(path)
		if err != nil {
			return "", fmt.Errorf("cannot read --kubeconfig file: %w", err)
		}
		defer file.Close()
		st, err := file.Stat()
		if err != nil || !st.Mode().IsRegular() {
			return "", fmt.Errorf("--kubeconfig must name a readable regular file")
		}
		return path, nil
	}
	if environment != "" {
		return "", nil
	}
	if home, err := os.UserHomeDir(); err == nil {
		if _, err := os.Stat(filepath.Join(home, ".kube/config")); !os.IsNotExist(err) {
			return "", nil
		}
	}
	if file, err := os.Open(k3s); err == nil {
		defer file.Close()
		if st, err := file.Stat(); err == nil && st.Mode().IsRegular() {
			return k3s, nil
		}
	}
	return "", nil
}
