// Agent Framework - Token Commands (Token Compression)
// Copyright (C) 2025 Agent Framework Contributors
//
// SPDX-License-Identifier: AGPL-3.0-or-later

package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

// tokenCmd represents the token command
var tokenCmd = &cobra.Command{
	Use:     "token",
	Aliases: []string{"tokens", "compress"},
	Short:   "Token 压缩与统计",
	Long: `管理和查看 Token 压缩统计信息，设置压缩策略。
Token 压缩可以将对话历史压缩到目标 Token 数量，节省 LLM API 开销。

注意：需要在配置中启用 tokenCompression 功能。

示例:
  af token stats                          # 查看压缩统计
  af token config                         # 查看压缩配置
  af token strategy truncate              # 设置压缩策略
  af token count "text to count"          # 统计文本 Token 数`,
}

// addTokenCommands adds token compression commands to root command
func addTokenCommands() {
	// ── stats ────────────────────────────────────────────────────────────────
	statsCmd := &cobra.Command{
		Use:   "stats",
		Short: "查看 Token 压缩统计",
		Long:  `显示 Token 压缩器的历史统计信息。`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			stats, err := app.GetHost().GetTokenCompressionStats()
			if err != nil {
				return fmt.Errorf("token compression is not configured or not enabled.\nAdd tokenCompression config to your host.yaml:\n  tokenCompression:\n    enabled: true\n    strategy: hybrid\n    targetTokens: 4000\n\nError: %w", err)
			}

			if outputFormat == "json" {
				b, _ := json.MarshalIndent(stats, "", "  ")
				fmt.Println(string(b))
				return nil
			}

			fmt.Println("Token Compression Statistics:")
			fmt.Println("────────────────────────────────────────────────────────────")
			fmt.Printf("Total Compressions:  %d\n", stats.TotalCompressions)
			fmt.Printf("Total Input Tokens:  %d\n", stats.TotalInputTokens)
			fmt.Printf("Total Output Tokens: %d\n", stats.TotalOutputTokens)
			fmt.Printf("Average Ratio:       %.2f%%\n", stats.AverageRatio*100)
			if stats.LastCompression != nil {
				fmt.Println()
				fmt.Println("Last Compression:")
				fmt.Printf("  Strategy:    %s\n", stats.LastCompression.Strategy)
				fmt.Printf("  Input:       %d tokens\n", stats.LastCompression.InputTokens)
				fmt.Printf("  Output:      %d tokens\n", stats.LastCompression.OutputTokens)
				fmt.Printf("  Ratio:       %.2f%%\n", stats.LastCompression.CompressionRatio*100)
				fmt.Printf("  Duration:    %dms\n", stats.LastCompression.DurationMs)
				success := "yes"
				if !stats.LastCompression.Success {
					success = fmt.Sprintf("no (%s)", stats.LastCompression.Error)
				}
				fmt.Printf("  Success:     %s\n", success)
			}
			fmt.Println("────────────────────────────────────────────────────────────")
			return nil
		},
	}
	tokenCmd.AddCommand(statsCmd)

	// ── config ───────────────────────────────────────────────────────────────
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "查看 Token 压缩配置",
		Long:  `显示当前 Token 压缩器的配置参数。`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := app.GetHost().Config()
			if cfg.TokenCompression == nil {
				fmt.Println("Token compression is not configured.")
				fmt.Println()
				fmt.Println("To enable it, add to your host.yaml:")
				fmt.Println("  tokenCompression:")
				fmt.Println("    enabled: true")
				fmt.Println("    strategy: hybrid")
				fmt.Println("    targetTokens: 4000")
				fmt.Println("    minTokens: 100")
				fmt.Println("    maxTokens: 16000")
				fmt.Println("    preserveSystemMessages: true")
				return nil
			}

			tc := cfg.TokenCompression

			if outputFormat == "json" {
				b, _ := json.MarshalIndent(tc, "", "  ")
				fmt.Println(string(b))
				return nil
			}

			fmt.Println("Token Compression Configuration:")
			fmt.Println("────────────────────────────────────────────────────────────")
			fmt.Printf("Enabled:                 %v\n", tc.Enabled)
			fmt.Printf("Strategy:                %s\n", tc.Strategy)
			fmt.Printf("Target Tokens:           %d\n", tc.TargetTokens)
			fmt.Printf("Min Tokens:              %d\n", tc.MinTokens)
			fmt.Printf("Max Tokens:              %d\n", tc.MaxTokens)
			fmt.Printf("Preserve System Msgs:    %v\n", tc.PreserveSystemMessages)
			fmt.Printf("Summary Model:           %s\n", tc.SummaryModelName)
			fmt.Printf("Summary Max Tokens:      %d\n", tc.SummaryMaxTokens)
			fmt.Printf("Temperature:             %.2f\n", tc.Temperature)
			fmt.Println("────────────────────────────────────────────────────────────")
			return nil
		},
	}
	tokenCmd.AddCommand(configCmd)

	// ── strategy ─────────────────────────────────────────────────────────────
	strategyCmd := &cobra.Command{
		Use:       "strategy [strategy-name]",
		Short:     "查看或设置压缩策略",
		Long:      `查看或设置 Token 压缩策略。支持的策略: truncate, summarize, hybrid, sliding_window。`,
		Args:      cobra.MaximumNArgs(1),
		ValidArgs: []string{"truncate", "summarize", "hybrid", "sliding_window"},
		RunE: func(cmd *cobra.Command, args []string) error {
			// No args: show current strategy
			if len(args) == 0 {
				cfg := app.GetHost().Config()
				if cfg.TokenCompression == nil {
					fmt.Println("Token compression is not configured")
					return nil
				}
				fmt.Printf("Current strategy: %s\n", cfg.TokenCompression.Strategy)
				fmt.Println()
				fmt.Println("Available strategies:")
				fmt.Println("  truncate        - Remove oldest messages")
				fmt.Println("  summarize       - Summarize with LLM")
				fmt.Println("  hybrid          - Truncate then summarize")
				fmt.Println("  sliding_window  - Keep recent context window")
				return nil
			}

			// Set strategy
			strategy := args[0]
			if err := app.GetHost().SetTokenCompressionStrategy(strategy); err != nil {
				return fmt.Errorf("failed to set strategy: %w", err)
			}

			fmt.Printf("✓ Token compression strategy set to: %s\n", strategy)
			return nil
		},
	}
	tokenCmd.AddCommand(strategyCmd)

	// ── count ────────────────────────────────────────────────────────────────
	countCmd := &cobra.Command{
		Use:   "count [text]",
		Short: "统计文本 Token 数量",
		Long:  `估算指定文本的 Token 数量（使用内置 Token 计数器）。`,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Join all args as text
			text := ""
			for i, arg := range args {
				if i > 0 {
					text += " "
				}
				text += arg
			}

			count := app.GetHost().CountTextTokens(text)
			fmt.Printf("Text: %q\n", truncateStr(text, 80))
			fmt.Printf("Token Count: %d\n", count)
			return nil
		},
	}
	tokenCmd.AddCommand(countCmd)

	// ── compress-text ─────────────────────────────────────────────────────────
	compressTextCmd := &cobra.Command{
		Use:   "compress-text [target-tokens] [text...]",
		Short: "压缩文本到目标 Token 数",
		Long:  `使用压缩器将提供的文本压缩到目标 Token 数量。`,
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			var targetTokens int
			if _, err := fmt.Sscanf(args[0], "%d", &targetTokens); err != nil {
				return fmt.Errorf("invalid target tokens '%s': must be a number", args[0])
			}

			text := ""
			for i, arg := range args[1:] {
				if i > 0 {
					text += " "
				}
				text += arg
			}

			ctx := rootContext()
			compressed, err := app.GetHost().CompressText(ctx, text, targetTokens)
			if err != nil {
				return fmt.Errorf("failed to compress text: %w", err)
			}

			originalCount := app.GetHost().CountTextTokens(text)
			compressedCount := app.GetHost().CountTextTokens(compressed)

			fmt.Printf("Original:   %d tokens\n", originalCount)
			fmt.Printf("Compressed: %d tokens\n", compressedCount)
			fmt.Printf("Ratio:      %.1f%%\n", float64(compressedCount)/float64(originalCount)*100)
			fmt.Println()
			fmt.Println("Compressed text:")
			fmt.Println("────────────────────────────────────────────────────────────")
			fmt.Println(compressed)
			return nil
		},
	}
	tokenCmd.AddCommand(compressTextCmd)

	rootCmd.AddCommand(tokenCmd)
}

// truncateStr truncates a string to maxLen with "..." suffix
func truncateStr(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
