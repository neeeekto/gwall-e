// Smoke spec proving the mockery v3 codegen + Gomega wiring (SVC-06, part of SC4).
// This is a throwaway unit spec against the generated ExampleProvisioner mock; real
// domain ports (and their specs) arrive in Phase 6/7. No build tag — runs as a unit
// test under pre-push.
package example_test

import (
	"context"
	"testing"

	"github.com/gwall-e/services/inventory/internal/example"
	"github.com/gwall-e/services/inventory/internal/example/mocks"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock" // regular qualified import (NOT dot, unlike ginkgo/gomega)
)

func TestExampleSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Example Package Suite")
}

var _ = Describe("ExampleProvisioner mock (mockery smoke)", func() {
	var (
		ctx context.Context
		m   *mocks.MockExampleProvisioner
	)

	BeforeEach(func() {
		ctx = context.Background()
		// NewMockX(GinkgoT()) registers an auto-Cleanup that asserts expectations.
		m = mocks.NewMockExampleProvisioner(GinkgoT())
	})

	Context("when the provision call succeeds", func() {
		BeforeEach(func() {
			m.EXPECT().Provision(mock.Anything, mock.Anything, mock.Anything).Return(nil)
		})

		It("returns no error", func() {
			err := m.Provision(ctx, example.ExampleID("ex-1"), "demo")
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("when the provision call is configured to fail", func() {
		BeforeEach(func() {
			m.EXPECT().
				Provision(mock.Anything, mock.Anything, mock.Anything).
				Return(example.ErrExampleProvisionFailed)
		})

		It("returns the sentinel error", func() {
			err := m.Provision(ctx, example.ExampleID("ex-2"), "demo")
			Expect(err).To(MatchError(example.ErrExampleProvisionFailed))
		})
	})
})
