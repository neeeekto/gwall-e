package http

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestIdmServicesSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "HTTP Package Suite")
}
