package gardenrunc_test

import (
	. "github.com/julz/garden-runc"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Create", func() {
	// var creator *DaemonContainerCreator
	// var depot *fakes.FakeDepot
	// var dockerRunner *fakes.FakeDockerRunner

	// BeforeEach(func() {
	// 	dockerRunner = new(fakes.FakeDockerRunner)
	// 	depot = new(fakes.FakeDepot)

	// 	depot.CreateReturns("the-depot-dir", nil)
	// })

	// JustBeforeEach(func() {
	// 	creator = &DaemonContainerCreator{
	// 		Depot:         depot,
	// 		InitdPath:     "bin-path/initd",
	// 		DoshPath:      "dosh-path",
	// 		DockerRunner:  dockerRunner,
	// 		DefaultRootfs: "docker:///thedefaultimage",
	// 	}
	// })

	// Context("when create is called", func() {
	// 	var createdContainer *Container
	// 	var createError error
	// 	var rootfsPath string

	// 	BeforeEach(func() {
	// 		rootfsPath = "docker:///somebuntu"
	// 	})

	// 	JustBeforeEach(func() {
	// 		createdContainer, createError = creator.Create(garden.ContainerSpec{
	// 			RootFSPath: rootfsPath,
	// 		})
	// 	})

	// 	It("creates a depot directory for the container", func() {
	// 		Expect(depot.CreateCallCount()).To(Equal(1))
	// 	})

	// 	Context("when creating the depot dir fails", func() {
	// 		BeforeEach(func() {
	// 			depot.CreateReturns("", errors.New("no depot for you"))
	// 		})

	// 		It("aborts the container creation", func() {
	// 			Expect(createError).To(MatchError("create depot dir: no depot for you"))
	// 			Expect(dockerRunner.RunCallCount()).To(Equal(0))
	// 		})
	// 	})

	// 	Context("when the rootfspath is not a url", func() {
	// 		BeforeEach(func() {
	// 			rootfsPath = "docker://%20/foo"
	// 		})

	// 		It("aborts the container creation", func() {
	// 			Expect(createError).To(MatchError("create: not a valid rootfs path: parse docker://%20/foo: hexadecimal escape in host"))
	// 			Expect(dockerRunner.RunCallCount()).To(Equal(0))
	// 		})
	// 	})

	// 	Context("and the docker run command fails", func() {
	// 		BeforeEach(func() {
	// 			dockerRunner.RunReturns("", errors.New("docker docker docker"))
	// 		})

	// 		It("returns a descriptive error", func() {
	// 			Expect(createError).To(MatchError("create: docker docker docker"))
	// 		})
	// 	})

	// 	PContext("andthe docker inspect command fails", func() {
	// 		BeforeEach(func() {
	// 			dockerRunner.InspectStub = func(cmd dockercli.InspectCmd) (string, error) {
	// 				return "", errors.New("something")
	// 			}
	// 		})

	// 		It("returns an error", func() {
	// 			Expect(createError).To(HaveOccurred())
	// 		})
	// 	})

	// 	Context("and the command succeeds", func() {
	// 		JustBeforeEach(func() {
	// 			Expect(createError).ToNot(HaveOccurred())
	// 		})

	// 		It("spawns initd inside a docker container", func() {
	// 			Expect(dockerRunner.RunArgsForCall(0).Program).To(Equal("/garden-bin/initd"))
	// 		})

	// 		It("asks for the image contained in the rootfspath", func() {
	// 			Expect(dockerRunner.RunArgsForCall(0).Image).To(Equal("somebuntu"))
	// 		})

	// 		It("tells docker to detach (to avoid blocking forever)", func() {
	// 			Expect(dockerRunner.RunArgsForCall(0).Detach).To(Equal(true))
	// 		})

	// 		Context("when the rootfspath is empty", func() {
	// 			BeforeEach(func() {
	// 				rootfsPath = ""
	// 			})

	// 			It("uses the default rootfspath", func() {
	// 				Expect(dockerRunner.RunArgsForCall(0).Image).To(Equal("thedefaultimage"))
	// 			})
	// 		})

	// 		It("mounts the initd executable into the container", func() {
	// 			Expect(dockerRunner.RunArgsForCall(0).Volumes).To(
	// 				ContainElement(dockercli.Volume{
	// 					HostPath:      "bin-path",
	// 					ContainerPath: "/garden-bin",
	// 				}),
	// 			)
	// 		})

	// 		It("mounts the ./run directory into the container", func() {
	// 			Expect(dockerRunner.RunArgsForCall(0).Volumes).To(
	// 				ContainElement(dockercli.Volume{
	// 					HostPath:      "the-depot-dir/run",
	// 					ContainerPath: "/run",
	// 				}),
	// 			)
	// 		})

	// 		It("tells initd to listen on /run/initd.sock and unmount /run afterwards", func() {
	// 			Expect(dockerRunner.RunArgsForCall(0).ProgramArgs).To(
	// 				Equal([]string{
	// 					"-socketPath", "/run/initd.sock",
	// 					"-unmountDir", "/run",
	// 				}),
	// 			)
	// 		})

	// 		Describe("the created container", func() {
	// 			BeforeEach(func() {
	// 				dockerRunner.RunReturns("docker-container-id", nil)
	// 				dockerRunner.InspectStub = func(cmd dockercli.InspectCmd) (string, error) {
	// 					return cmd.Field + " of " + cmd.ContainerID, nil
	// 				}
	// 			})

	// 			It("is configured to run commands via dosh", func() {
	// 				cmd := createdContainer.ContainerCmd.Cmd("foo", "bar", "baz")

	// 				Expect(cmd.Path).To(Equal("dosh-path"))
	// 				Expect(cmd.Args).To(Equal([]string{
	// 					"dosh-path",
	// 					"-socketPath", "the-depot-dir/run/initd.sock",
	// 					"foo", "bar", "baz",
	// 				}))
	// 			})

	// 			It("has its containerPath set", func() {
	// 				Expect(createdContainer.InfoHandler.ContainerPath).To(Equal("the-depot-dir"))
	// 			})

	// 			It("has its ContainerIP set (based on the output of the docker inspect command)", func() {
	// 				Expect(createdContainer.InfoHandler.ContainerIP).To(Equal("NetworkSettings.IPAddress of docker-container-id"))
	// 			})

	// 			It("has its docker id set", func() {
	// 				Expect(createdContainer.InfoHandler.DockerID).To(Equal("docker-container-id"))
	// 			})
	// 		})
	// 	})
	// })
})
