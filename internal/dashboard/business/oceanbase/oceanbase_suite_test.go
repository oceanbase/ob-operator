package oceanbase_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestOceanbase(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Oceanbase Suite")
}
