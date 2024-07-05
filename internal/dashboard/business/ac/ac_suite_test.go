package ac_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestAc(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Ac Suite")
}
