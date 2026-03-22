// Agent Framework - Plugin Commands
// Copyright (C) 2025 Agent Framework Contributors
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package cli

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

// pluginCmd represents the plugin command
var pluginCmd = &cobra.Command{
	Use:     "plugin",
	Aliases: []string{"plugins", "pl"},
	Short:   "管理插件",
	Long: `管理 AgentFramework 插件的加载、启用、禁用和卸载。
插件是扩展框架功能的模块，可以提供新的技能、工作流、模型适配器等功能。

示例:
  af plugin list                     # 列出所有插件
  af plugin info my-plugin           # 查看插件详情
  af plugin load /path/to/plugin     # 加载插件
  af plugin enable my-plugin         # 启用插件
  af plugin disable my-plugin        # 禁用插件
  af plugin unload my-plugin         # 卸载插件`,
}

// addPluginCommands adds plugin-related commands to root command
func addPluginCommands() {
	// ── list ─────────────────────────────────────────────────────────────────
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "列出所有插件",
		Long:  `列出当前已注册的所有插件及其状态。`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			pluginMgr := app.GetHost().PluginManager()
			if pluginMgr == nil {
				fmt.Println("No plugin manager available")
				return nil
			}

			plugins := pluginMgr.GetAllPlugins()

			if outputFormat == "json" {
				type pluginJSON struct {
					Name        string `json:"name"`
					Version     string `json:"version"`
					Description string `json:"description"`
					Enabled     bool   `json:"enabled"`
				}
				list := make([]pluginJSON, 0, len(plugins))
				for _, p := range plugins {
					list = append(list, pluginJSON{
						Name:        p.Name(),
						Version:     p.Version(),
						Description: p.Description(),
						Enabled:     p.IsEnabled(),
					})
				}
				b, _ := json.MarshalIndent(list, "", "  ")
				fmt.Println(string(b))
				return nil
			}

			if len(plugins) == 0 {
				fmt.Println("No plugins registered")
				fmt.Println()
				fmt.Println("To load a plugin, use:")
				fmt.Println("  af plugin load /path/to/plugin")
				return nil
			}

			fmt.Println("Plugins:")
			fmt.Println("────────────────────────────────────────────────────────────")
			for _, p := range plugins {
				status := "disabled"
				if p.IsEnabled() {
					status = "enabled"
				}
				fmt.Printf("  %-20s  v%-10s  %-8s  %s\n",
					p.Name(), p.Version(), status, p.Description())
			}
			fmt.Printf("Total: %d plugin(s)\n", len(plugins))
			return nil
		},
	}
	pluginCmd.AddCommand(listCmd)

	// ── info ─────────────────────────────────────────────────────────────────
	infoCmd := &cobra.Command{
		Use:   "info [plugin-name]",
		Short: "查看插件详情",
		Long:  `查看指定插件的详细信息。`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pluginMgr := app.GetHost().PluginManager()
			if pluginMgr == nil {
				return fmt.Errorf("plugin manager not available")
			}

			p, ok := pluginMgr.GetPlugin(args[0])
			if !ok {
				return fmt.Errorf("plugin '%s' not found", args[0])
			}

			if outputFormat == "json" {
				info := map[string]interface{}{
					"name":        p.Name(),
					"version":     p.Version(),
					"description": p.Description(),
					"enabled":     p.IsEnabled(),
				}
				b, _ := json.MarshalIndent(info, "", "  ")
				fmt.Println(string(b))
				return nil
			}

			fmt.Println("Plugin Information:")
			fmt.Println("────────────────────────────────────────────────────────────")
			fmt.Printf("Name:        %s\n", p.Name())
			fmt.Printf("Version:     %s\n", p.Version())
			fmt.Printf("Description: %s\n", p.Description())
			enabled := "No"
			if p.IsEnabled() {
				enabled = "Yes"
			}
			fmt.Printf("Enabled:     %s\n", enabled)
			fmt.Println("────────────────────────────────────────────────────────────")
			return nil
		},
	}
	pluginCmd.AddCommand(infoCmd)

	// ── load ─────────────────────────────────────────────────────────────────
	loadCmd := &cobra.Command{
		Use:   "load [plugin-path]",
		Short: "加载插件",
		Long:  `从指定路径加载一个插件。`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := rootContext()
			pluginMgr := app.GetHost().PluginManager()
			if pluginMgr == nil {
				return fmt.Errorf("plugin manager not available")
			}

			pluginPath := args[0]
			if err := pluginMgr.LoadPlugin(ctx, pluginPath); err != nil {
				return fmt.Errorf("failed to load plugin from '%s': %w", pluginPath, err)
			}

			fmt.Printf("✓ Plugin loaded from: %s\n", pluginPath)
			return nil
		},
	}
	pluginCmd.AddCommand(loadCmd)

	// ── load-dir ─────────────────────────────────────────────────────────────
	loadDirCmd := &cobra.Command{
		Use:   "load-dir [directory]",
		Short: "从目录批量加载插件",
		Long:  `扫描指定目录并加载所有找到的插件。`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := rootContext()
			pluginMgr := app.GetHost().PluginManager()
			if pluginMgr == nil {
				return fmt.Errorf("plugin manager not available")
			}

			dir := args[0]
			if err := pluginMgr.LoadPluginsFromDirectory(ctx, dir); err != nil {
				return fmt.Errorf("failed to load plugins from directory '%s': %w", dir, err)
			}

			fmt.Printf("✓ Plugins loaded from directory: %s\n", dir)
			return nil
		},
	}
	pluginCmd.AddCommand(loadDirCmd)

	// ── enable ───────────────────────────────────────────────────────────────
	enableCmd := &cobra.Command{
		Use:   "enable [plugin-name...]",
		Short: "启用插件",
		Long:  `启用一个或多个插件。`,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := rootContext()
			pluginMgr := app.GetHost().PluginManager()
			if pluginMgr == nil {
				return fmt.Errorf("plugin manager not available")
			}

			for _, name := range args {
				if err := pluginMgr.EnablePlugin(ctx, name); err != nil {
					return fmt.Errorf("failed to enable plugin '%s': %w", name, err)
				}
				fmt.Printf("✓ Plugin '%s' enabled\n", name)
			}
			return nil
		},
	}
	pluginCmd.AddCommand(enableCmd)

	// ── disable ──────────────────────────────────────────────────────────────
	disableCmd := &cobra.Command{
		Use:   "disable [plugin-name...]",
		Short: "禁用插件",
		Long:  `禁用一个或多个插件（不会卸载）。`,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := rootContext()
			pluginMgr := app.GetHost().PluginManager()
			if pluginMgr == nil {
				return fmt.Errorf("plugin manager not available")
			}

			for _, name := range args {
				if err := pluginMgr.DisablePlugin(ctx, name); err != nil {
					return fmt.Errorf("failed to disable plugin '%s': %w", name, err)
				}
				fmt.Printf("✓ Plugin '%s' disabled\n", name)
			}
			return nil
		},
	}
	pluginCmd.AddCommand(disableCmd)

	// ── unload ───────────────────────────────────────────────────────────────
	unloadCmd := &cobra.Command{
		Use:     "unload [plugin-name]",
		Aliases: []string{"remove", "rm"},
		Short:   "卸载插件",
		Long:    `卸载并移除指定插件（会调用 Shutdown 并从注册表删除）。`,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := rootContext()
			pluginMgr := app.GetHost().PluginManager()
			if pluginMgr == nil {
				return fmt.Errorf("plugin manager not available")
			}

			if err := pluginMgr.UnloadPlugin(ctx, args[0]); err != nil {
				return fmt.Errorf("failed to unload plugin '%s': %w", args[0], err)
			}

			fmt.Printf("✓ Plugin '%s' unloaded\n", args[0])
			return nil
		},
	}
	pluginCmd.AddCommand(unloadCmd)

	// ── reload ───────────────────────────────────────────────────────────────
	reloadCmd := &cobra.Command{
		Use:   "reload [plugin-name]",
		Short: "重新加载插件",
		Long:  `卸载后重新加载指定插件（热更新）。`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := rootContext()
			pluginMgr := app.GetHost().PluginManager()
			if pluginMgr == nil {
				return fmt.Errorf("plugin manager not available")
			}

			if err := pluginMgr.ReloadPlugin(ctx, args[0]); err != nil {
				return fmt.Errorf("failed to reload plugin '%s': %w", args[0], err)
			}

			fmt.Printf("✓ Plugin '%s' reloaded\n", args[0])
			return nil
		},
	}
	pluginCmd.AddCommand(reloadCmd)

	// ── init-all ─────────────────────────────────────────────────────────────
	initAllCmd := &cobra.Command{
		Use:   "init-all",
		Short: "初始化所有插件",
		Long:  `对所有已加载的插件调用 Initialize 生命周期方法。`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			pluginMgr := app.GetHost().PluginManager()
			if pluginMgr == nil {
				return fmt.Errorf("plugin manager not available")
			}

			if err := pluginMgr.InitializeAllPlugins(ctx, app.GetHost()); err != nil {
				return fmt.Errorf("failed to initialize plugins: %w", err)
			}

			fmt.Println("✓ All plugins initialized")
			return nil
		},
	}
	pluginCmd.AddCommand(initAllCmd)

	rootCmd.AddCommand(pluginCmd)
}
