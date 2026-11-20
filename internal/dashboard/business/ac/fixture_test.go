package ac

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Ginkgo business tests must not depend on or mutate a developer's policy file.
var _ = BeforeSuite(func() {
	directory, err := os.MkdirTemp("", "dashboard-ac-test-")
	Expect(err).NotTo(HaveOccurred())
	previous := enforcer
	DeferCleanup(func() { enforcer = previous; _ = os.RemoveAll(directory) })
	path := filepath.Join(directory, "rbac_policy.csv")
	policy := "p, admin, *, *, \"Super admin\"\np, admin2, book/*, read, \"Book reader\"\ng, admin, admin\ng, admin, admin2\n"
	Expect(os.WriteFile(path, []byte(policy), 0600)).To(Succeed())
	enforcer, err = initEnforcer(path, "test-only-policy", "rbac_policy.csv", "test-only")
	Expect(err).NotTo(HaveOccurred())
})
