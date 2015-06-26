package gardenrunc_test

import (
	"io/ioutil"
	"path"

	. "github.com/julz/garden-runc"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Depot", func() {
	var depot *ContainerDepot

	BeforeEach(func() {
		tmp, err := ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())

		depot = &ContainerDepot{
			Dir: tmp,
		}
	})

	Describe("Create", func() {
		It("creates a run directory to store the initd socket", func() {
			dir, err := depot.Create()
			Expect(err).NotTo(HaveOccurred())
			Expect(path.Join(dir, "run")).To(BeADirectory())
		})

		It("creates a processes directory to store the iodaemon sockets", func() {
			dir, err := depot.Create()
			Expect(err).NotTo(HaveOccurred())
			Expect(path.Join(dir, "processes")).To(BeADirectory())
		})

		It("copies iodaemon into bin/ directory", func() {
			dir, err := depot.Create()
			Expect(err).NotTo(HaveOccurred())
			Expect(path.Join(dir, "bin", "iodaemon")).To(BeAnExistingFile())
		})

		Context("with multiple containers", func() {
			It("creates distinct run directories for each container", func() {
				dir1, err := depot.Create()
				Expect(err).NotTo(HaveOccurred())

				dir2, err := depot.Create()
				Expect(err).NotTo(HaveOccurred())

				Expect(dir2).NotTo(Equal(dir1))
			})
		})
	})
})
