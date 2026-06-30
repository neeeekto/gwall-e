// Ginkgo suite bootstrap for the inventory domain package (testing.md canon).
package domain_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestDomainSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Inventory Domain Suite")
}
