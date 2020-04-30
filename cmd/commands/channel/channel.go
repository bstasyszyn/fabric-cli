/*
Copyright State Street Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package channel

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/hyperledger/fabric-cli/cmd/common"
	"github.com/hyperledger/fabric-cli/pkg/environment"
	"github.com/hyperledger/fabric-cli/pkg/fabric"
)

// NewChannelCommand creates a new "fabric channel" command
func NewChannelCommand(settings *environment.Settings) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "channel",
		Short: "Manage channels",
	}

	cmd.AddCommand(
		NewChannelCreateCommand(settings),
		NewChannelJoinCommand(settings),
		NewChannelUpdateCommand(settings),
		NewChannelListCommand(settings),
		NewChannelConfigCommand(settings),
	)

	cmd.SetOutput(settings.Streams.Out)

	return cmd
}

// BaseCommand implements common channel command functions
type BaseCommand struct {
	common.Command

	Factory            fabric.Factory
	ResourceManagement fabric.ResourceManagement
}

// Complete initializes all clients needed for Run
func (c *BaseCommand) Complete() error {
	var err error

	if c.Factory == nil {
		c.Factory, err = fabric.NewFactory(c.Settings.Config)
		if err != nil {
			return err
		}
	}

	c.ResourceManagement, err = c.Factory.ResourceManagement()
	if err != nil {
		return err
	}

	go c.closeOnExit()

	return nil
}

func (c *BaseCommand) closeOnExit() {
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigs
		fmt.Println(sig)
		done <- true
	}()

	fmt.Println("awaiting signal...")

	<-done

	fmt.Println("... exiting")

	if c.Factory != nil {
		sdk, err := c.Factory.SDK()
		if err != nil {
			fmt.Println(err.Error())
		} else {
			fmt.Println("Closing SDK")
			sdk.Close()
		}
	}
}
