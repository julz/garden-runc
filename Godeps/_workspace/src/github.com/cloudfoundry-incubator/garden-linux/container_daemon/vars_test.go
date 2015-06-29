package container_daemon_test

import (
	. "github.com/cloudfoundry-incubator/garden-linux/container_daemon"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Vars", func() {
	Describe("when set is called", func() {

		var sl *StringList

		BeforeEach(func() {
			sl = &StringList{}
			Expect(sl.Set("foo")).To(Succeed())
			Expect(sl.Set("bar")).To(Succeed())
		})

		It("adds the value to the list", func() {
			Expect(sl.List).To(ConsistOf("foo", "bar"))
		})

		It("stringifies with commas", func() {
			Expect(sl.String()).To(Equal("foo, bar"))
		})

		It("returns the list from Get()", func() {
			Expect(sl.Get()).To(Equal([]string{"foo", "bar"}))
		})
	})
})
