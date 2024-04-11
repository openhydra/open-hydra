package train_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestTrain(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Train Suite")
}
