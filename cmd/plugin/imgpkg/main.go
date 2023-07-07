package main

import (
	"encoding/json"
	"fmt"
	"os"

	uierrs "github.com/cppforlife/go-cli-ui/errors"
	"github.com/cppforlife/go-cli-ui/ui"
	"github.com/spf13/cobra"
	"github.com/vmware-tanzu/carvel-imgpkg/pkg/imgpkg/cmd"
	"github.com/vmware-tanzu/tanzu-plugin-runtime/config/types"
	"github.com/vmware-tanzu/tanzu-plugin-runtime/log"
	"github.com/vmware-tanzu/tanzu-plugin-runtime/plugin"
	"github.com/vmware-tanzu/tanzu-plugin-runtime/plugin/buildinfo"
)

var descriptor = plugin.PluginDescriptor{
	Name:        "imgpkg",
	Description: "imgpkg allows to store configuration and image references as oci artifacts (copy, describe, pull, push, tag, version)",
	Target:      types.TargetGlobal,
	Version:     buildinfo.Version,
	BuildSHA:    buildinfo.SHA,
	Group:       plugin.ManageCmdGroup, // set group
}

func main() {
	p, err := plugin.NewPlugin(&descriptor)
	if err != nil {
		log.Fatal(err, "")
	}

	confUI := ui.NewConfUI(ui.NewNoopLogger())
	defer confUI.Flush()

	p.Cmd = cmd.NewDefaultImgpkgCmd(confUI)

	p.AddCommands(
		// required for building plugin
		newDescribeCmd(descriptor.Description),
		newInfoCmd(&descriptor),

		//required to initilize the plugin post installation
		newPostInstallCmd(&descriptor),
	)

	if err := p.Execute(); err != nil {
		confUI.ErrorLinef("imgpkg: Error: %v", uierrs.NewMultiLineError(err))
		os.Exit(1)
	}
}

func newDescribeCmd(description string) *cobra.Command {
	cmd := &cobra.Command{
		Use:    "describe",
		Short:  "Describes the plugin",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println(description)
			return nil
		},
	}

	return cmd
}

func newInfoCmd(desc *plugin.PluginDescriptor) *cobra.Command {
	cmd := &cobra.Command{
		Use:    "info",
		Short:  "Plugin info",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			pi := pluginInfo{
				PluginDescriptor:     *desc,
				PluginRuntimeVersion: "test",
			}
			b, err := json.Marshal(pi)
			if err != nil {
				return err
			}
			fmt.Println(string(b))
			return nil
		},
	}

	return cmd
}

func newPostInstallCmd(desc *plugin.PluginDescriptor) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "post-install",
		Short:        "Run post install configuration for a plugin",
		Long:         "Run post install configuration for a plugin",
		Hidden:       true,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			// invoke postInstall for the plugin
			return desc.PostInstallHook()
		},
	}

	return cmd
}

// pluginInfo describes a plugin information. This is a super set of PluginDescriptor
// It includes some additional metadata that plugin runtime configures
type pluginInfo struct {
	// PluginDescriptor describes a plugin binary.
	plugin.PluginDescriptor `json:",inline" yaml:",inline"`

	// PluginRuntimeVersion of the plugin. Must be a valid semantic version https://semver.org/
	// This version specifies the version of Plugin Runtime that was used to build the plugin
	PluginRuntimeVersion string `json:"pluginRuntimeVersion" yaml:"pluginRuntimeVersion"`
}
