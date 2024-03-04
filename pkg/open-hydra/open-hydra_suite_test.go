package openhydra_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestOpenHydra(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "OpenHydra Suite")
}
