/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package lifecycle_test

import (
	"bytes"
	"errors"
	"fmt"
	"os"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"

	"github.com/hyperledger/fabric-cli/cmd/commands/lifecycle"
	"github.com/hyperledger/fabric-cli/pkg/environment"
	"github.com/hyperledger/fabric-cli/pkg/fabric/mocks"
)

var _ = Describe("LifecycleCommitCommand", func() {
	var (
		cmd      *cobra.Command
		settings *environment.Settings
		out      *bytes.Buffer

		args []string
	)

	BeforeEach(func() {
		out = new(bytes.Buffer)

		settings = &environment.Settings{
			Home: environment.Home(os.TempDir()),
			Streams: environment.Streams{
				Out: out,
			},
		}

		args = os.Args
	})

	JustBeforeEach(func() {
		cmd = lifecycle.NewCommitCommand(settings)
	})

	AfterEach(func() {
		os.Args = args
	})

	It("should create a lifecycle commit command", func() {
		Expect(cmd.Name()).To(Equal("commit"))
		Expect(cmd.HasSubCommands()).To(BeFalse())
	})

	It("should provide a help prompt", func() {
		os.Args = append(os.Args, "--help")

		Expect(cmd.Execute()).Should(Succeed())
		Expect(fmt.Sprint(out)).To(ContainSubstring("commit <chaincode-name> <version> <sequence>"))
	})
})

var _ = Describe("LifecycleCommitImplementation", func() {
	var (
		impl     *lifecycle.CommitCommand
		err      error
		out      *bytes.Buffer
		settings *environment.Settings
		factory  *mocks.Factory
		client   *mocks.ResourceManagement
	)

	BeforeEach(func() {
		out = new(bytes.Buffer)

		settings = &environment.Settings{
			Home: environment.Home(os.TempDir()),
			Streams: environment.Streams{
				Out: out,
			},
		}

		factory = &mocks.Factory{}
		client = &mocks.ResourceManagement{}

		impl = &lifecycle.CommitCommand{}
		impl.Settings = settings
		impl.Factory = factory
	})

	It("should not be nil", func() {
		Expect(impl).ShouldNot(BeNil())
	})

	Describe("Validate", func() {
		JustBeforeEach(func() {
			err = impl.Validate()
		})

		It("should fail when name is not set", func() {
			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(Equal("chaincode name not specified"))
		})

		Context("when chaincode version is not set", func() {
			BeforeEach(func() {
				impl.Name = "mycc"
			})

			It("should fail without chaincode version", func() {
				Expect(err).NotTo(BeNil())
				Expect(err.Error()).To(Equal("chaincode version not specified"))
			})
		})

		Context("when chaincode sequence is not set", func() {
			BeforeEach(func() {
				impl.Name = "mycc"
				impl.Version = "0.0.0"
			})

			It("should fail without chaincode sequence", func() {
				Expect(err).NotTo(BeNil())
				Expect(err.Error()).To(Equal("sequence not specified"))
			})
		})

		Context("when chaincode sequence is not greater than 0", func() {
			BeforeEach(func() {
				impl.Name = "mycc"
				impl.Version = "0.0.0"
				impl.Sequence = "-1"
			})

			It("should fail with chaincode sequence not greater than 0", func() {
				Expect(err).NotTo(BeNil())
				Expect(err.Error()).To(Equal("sequence must be greater than 0"))
			})
		})

		Context("when chaincode sequence is invalid", func() {
			BeforeEach(func() {
				impl.Name = "mycc"
				impl.Version = "0.0.0"
				impl.Sequence = "xxx"
			})

			It("should fail with chaincode sequence is invalid", func() {
				Expect(err).NotTo(BeNil())
				Expect(err.Error()).To(ContainSubstring("invalid sequence"))
			})
		})

		Context("when at least one peer is not specified", func() {
			BeforeEach(func() {
				impl.Name = "mycc"
				impl.Version = "0.0.0"
				impl.Sequence = "1"
			})

			It("should fail with at least one peer is not specified", func() {
				Expect(err.Error()).To(Equal("at least one peer must be specified"))
			})
		})

		Context("when all arguments are set", func() {
			BeforeEach(func() {
				impl.Name = "mycc"
				impl.Version = "0.0.0"
				impl.Sequence = "1"
				impl.Peers = []string{"peer1"}
			})

			It("should succeed with all arguments", func() {
				Expect(err).To(BeNil())
			})
		})
	})

	Describe("Run", func() {
		BeforeEach(func() {
			impl.Name = "mycc"
			impl.Version = "0.0.0"
			impl.Sequence = "1"
			impl.ResourceManagement = client
		})

		JustBeforeEach(func() {
			err = impl.Run()
		})

		It("should fail without a current context", func() {
			Expect(err).NotTo(BeNil())
		})

		Context("when resmgmt client succeeds", func() {
			BeforeEach(func() {
				settings.Config = &environment.Config{
					Contexts: map[string]*environment.Context{
						"foo": {},
					},
					CurrentContext: "foo",
				}

				client.InstantiateCCReturns(resmgmt.InstantiateCCResponse{}, nil)
			})

			It("should succeed with chaincode commit", func() {
				Expect(err).To(BeNil())
				Expect(fmt.Sprint(out)).To(Equal("successfully committed chaincode 'mycc'\n"))
			})
		})

		Context("when resmgmt client fails", func() {
			BeforeEach(func() {
				settings.Config = &environment.Config{
					Contexts: map[string]*environment.Context{
						"foo": {},
					},
					CurrentContext: "foo",
				}

				client.LifecycleCommitCCReturns(fab.TransactionID(""), errors.New("commit error"))
			})

			It("should fail to commit chaincode", func() {
				Expect(err).NotTo(BeNil())
				Expect(err.Error()).To(ContainSubstring("commit error"))
			})
		})
	})
})
