// Agent Framework - Config Commands
// Copyright (C) 2025 Agent Framework Contributors
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

package cmdcli

import (
	"fmt"

	"github.com/spf13/cobra"

	"AgentFramework/core"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "管理配置",
	Long:  `管理应用配置，包括查看、设置和重新加载配置。`,
}

// addConfigCommands adds config-related commands to root command
func addConfigCommands() {
	// Get config
	getCmd := &cobra.Command{
		Use:   "get",
		Short: "获取配置信息",
		Long:  `获取当前配置信息。可以指定配置键获取特定值。`,
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := rootContext()
			svc := core.NewConfigService(app)

			if len(args) == 0 {
				return svc.PrintConfig(ctx, outputFormat)
			}

			// Get specific config value
			value, err := svc.GetConfigValue(ctx, args[0])
			if err != nil {
				return err
			}

			fmt.Printf("%s: %v\n", args[0], value)
			return nil
		},
	}
	configCmd.AddCommand(getCmd)

	rootCmd.AddCommand(configCmd)
}
