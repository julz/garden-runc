package gardenrunc_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestGardenDocker(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "GardenDocker Suite")
}
