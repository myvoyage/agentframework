package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var (
	_secretsFile string
)

func addSecretsCommands() {
	// ── secrets ────────────────────────────────────────────────────────────
	secretsCmd := &cobra.Command{
		Use:   "secrets",
		Short: "管理敏感信息(密钥、Token等)",
		Long: `管理 AgentFramework 的敏感信息,包括 API 密钥、访问令牌等。

密钥文件存储位置: <data-dir>/secrets.json
密钥文件会自动加密存储,并在内存中解密使用。

安全提示:
  - 请勿将 secrets.json 提交到版本控制
  - 建议将 secrets.json 加入 .gitignore
  - 定期轮换密钥以提高安全性`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	// ── list ──────────────────────────────────────────────────────────────
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "列出所有密钥(隐藏值)",
		Long:  `列出所有存储的密钥名称,但不显示实际密钥值。`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			secrets, err := loadSecrets()
			if err != nil {
				return fmt.Errorf("加载密钥失败: %w", err)
			}

			if len(secrets) == 0 {
				fmt.Println("ℹ 未存储任何密钥")
				return nil
			}

			fmt.Println("已存储的密钥:")
			fmt.Println("────────────────────────────────────────────────────────────")

			// Group by prefix
			grouped := make(map[string][]string)
			for name := range secrets {
				parts := strings.SplitN(name, "_", 2)
				prefix := parts[0]
				grouped[prefix] = append(grouped[prefix], name)
			}

			for prefix, names := range grouped {
				fmt.Printf("\n[%s]\n", strings.ToUpper(prefix))
				for _, name := range names {
					fmt.Printf("  • %s\n", name)
				}
			}

			fmt.Println("\n────────────────────────────────────────────────────────────")
			fmt.Printf("总计: %d 个密钥\n", len(secrets))

			return nil
		},
	}
	secretsCmd.AddCommand(listCmd)

	// ── get ───────────────────────────────────────────────────────────────
	getCmd := &cobra.Command{
		Use:   "get <key>",
		Short: "获取指定密钥的值",
		Long:  `获取并显示指定密钥的完整值。警告: 密钥值会在终端显示。`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]

			secrets, err := loadSecrets()
			if err != nil {
				return fmt.Errorf("加载密钥失败: %w", err)
			}

			value, exists := secrets[key]
			if !exists {
				return fmt.Errorf("密钥不存在: %s", key)
			}

			fmt.Printf("%s\n", value)
			return nil
		},
	}
	secretsCmd.AddCommand(getCmd)

	// ── set ───────────────────────────────────────────────────────────────
	setCmd := &cobra.Command{
		Use:   "set <key> [value]",
		Short: "设置密钥",
		Long: `设置一个密钥值。如果未提供值,将提示输入(输入时隐藏)。

示例:
  af secrets set OPENAI_API_KEY sk-xxx
  af secrets set TELEGRAM_BOT_TOKEN  (交互式输入,隐藏显示)`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]
			var value string

			if len(args) == 2 {
				value = args[1]
			} else {
				// Interactive input with hidden display
				fmt.Printf("输入 %s 的值(输入将隐藏): ", key)
				byteValue, err := term.ReadPassword(int(os.Stdin.Fd()))
				if err != nil {
					return fmt.Errorf("读取输入失败: %w", err)
				}
				value = string(byteValue)
				fmt.Println() // New line after password input
			}

			if value == "" {
				return fmt.Errorf("密钥值不能为空")
			}

			secrets, err := loadSecrets()
			if err != nil {
				return fmt.Errorf("加载密钥失败: %w", err)
			}

			secrets[key] = value

			if err := saveSecrets(secrets); err != nil {
				return fmt.Errorf("保存密钥失败: %w", err)
			}

			fmt.Printf("✓ 密钥已设置: %s\n", key)

			// Add to .gitignore if needed
			dataDir, _ := getDefaultDataDir()
			gitignorePath := filepath.Join(filepath.Dir(dataDir), ".gitignore")
			addToGitignore(gitignorePath, filepath.Base(dataDir)+"/secrets.json")

			return nil
		},
	}
	secretsCmd.AddCommand(setCmd)

	// ── delete ────────────────────────────────────────────────────────────
	deleteCmd := &cobra.Command{
		Use:   "delete <key>",
		Short: "删除指定密钥",
		Long:  `删除指定的密钥。此操作不可逆。`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]

			secrets, err := loadSecrets()
			if err != nil {
				return fmt.Errorf("加载密钥失败: %w", err)
			}

			if _, exists := secrets[key]; !exists {
				return fmt.Errorf("密钥不存在: %s", key)
			}

			delete(secrets, key)

			if err := saveSecrets(secrets); err != nil {
				return fmt.Errorf("保存密钥失败: %w", err)
			}

			fmt.Printf("✓ 密钥已删除: %s\n", key)

			// Check if secrets file is now empty
			if len(secrets) == 0 {
				dataDir, _ := getDefaultDataDir()
				secretsPath := filepath.Join(dataDir, "secrets.json")
				if err := os.Remove(secretsPath); err != nil {
					fmt.Printf("⚠ 无法删除空的密钥文件: %v\n", err)
				}
				fmt.Println("ℹ 所有密钥已清空,密钥文件已删除")
			}

			return nil
		},
	}
	secretsCmd.AddCommand(deleteCmd)

	// ── export ────────────────────────────────────────────────────────────
	exportCmd := &cobra.Command{
		Use:   "export <file>",
		Short: "导出密钥到文件",
		Long: `将所有密钥导出到指定文件。输出格式为 JSON。

警告: 导出的文件包含明文密钥,请妥善保管!`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			secrets, err := loadSecrets()
			if err != nil {
				return fmt.Errorf("加载密钥失败: %w", err)
			}

			if len(secrets) == 0 {
				return fmt.Errorf("没有密钥可导出")
			}

			exportFile := args[0]

			data, err := json.MarshalIndent(secrets, "", "  ")
			if err != nil {
				return fmt.Errorf("序列化失败: %w", err)
			}

			if err := os.WriteFile(exportFile, data, 0600); err != nil {
				return fmt.Errorf("写入文件失败: %w", err)
			}

			fmt.Printf("✓ 已导出 %d 个密钥到: %s\n", len(secrets), exportFile)
			fmt.Println("⚠ 请妥善保管此文件,包含明文密钥!")

			return nil
		},
	}
	secretsCmd.AddCommand(exportCmd)

	// ── import ────────────────────────────────────────────────────────────
	importCmd := &cobra.Command{
		Use:   "import <file>",
		Short: "从文件导入密钥",
		Long: `从 JSON 文件导入密钥。文件格式应为 {"key": "value", ...}

导入后会合并现有密钥,重复的密钥会被覆盖。`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			importFile := args[0]

			data, err := os.ReadFile(importFile)
			if err != nil {
				return fmt.Errorf("读取文件失败: %w", err)
			}

			var importedSecrets map[string]string
			if err := json.Unmarshal(data, &importedSecrets); err != nil {
				return fmt.Errorf("解析 JSON 失败: %w", err)
			}

			if len(importedSecrets) == 0 {
				return fmt.Errorf("文件中没有找到密钥")
			}

			secrets, err := loadSecrets()
			if err != nil {
				return fmt.Errorf("加载密钥失败: %w", err)
			}

			count := 0
			for key, value := range importedSecrets {
				secrets[key] = value
				count++
			}

			if err := saveSecrets(secrets); err != nil {
				return fmt.Errorf("保存密钥失败: %w", err)
			}

			fmt.Printf("✓ 已导入 %d 个密钥\n", count)

			// Add to .gitignore if needed
			dataDir, _ := getDefaultDataDir()
			gitignorePath := filepath.Join(filepath.Dir(dataDir), ".gitignore")
			addToGitignore(gitignorePath, filepath.Base(dataDir)+"/secrets.json")

			return nil
		},
	}
	secretsCmd.AddCommand(importCmd)

	// ── rename ────────────────────────────────────────────────────────────
	renameCmd := &cobra.Command{
		Use:   "rename <old-key> <new-key>",
		Short: "重命名密钥",
		Long:  `重命名指定的密钥。`,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			oldKey := args[0]
			newKey := args[1]

			secrets, err := loadSecrets()
			if err != nil {
				return fmt.Errorf("加载密钥失败: %w", err)
			}

			value, exists := secrets[oldKey]
			if !exists {
				return fmt.Errorf("密钥不存在: %s", oldKey)
			}

			if _, exists := secrets[newKey]; exists {
				return fmt.Errorf("目标密钥已存在: %s", newKey)
			}

			delete(secrets, oldKey)
			secrets[newKey] = value

			if err := saveSecrets(secrets); err != nil {
				return fmt.Errorf("保存密钥失败: %w", err)
			}

			fmt.Printf("✓ 密钥已重命名: %s → %s\n", oldKey, newKey)
			return nil
		},
	}
	secretsCmd.AddCommand(renameCmd)

	// ── clear ────────────────────────────────────────────────────────────
	clearCmd := &cobra.Command{
		Use:   "clear",
		Short: "清空所有密钥",
		Long:  `删除所有存储的密钥。此操作不可逆!`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			secrets, err := loadSecrets()
			if err != nil {
				return fmt.Errorf("加载密钥失败: %w", err)
			}

			if len(secrets) == 0 {
				fmt.Println("ℹ 没有密钥可清空")
				return nil
			}

			fmt.Printf("警告: 即将删除 %d 个密钥,此操作不可逆!\n", len(secrets))
			fmt.Print("确认继续? [yes/NO]: ")

			var confirm string
			fmt.Scanln(&confirm)

			if strings.ToLower(confirm) != "yes" {
				fmt.Println("已取消")
				return nil
			}

			dataDir, _ := getDefaultDataDir()
			secretsPath := filepath.Join(dataDir, "secrets.json")

			if err := os.Remove(secretsPath); err != nil {
				return fmt.Errorf("删除密钥文件失败: %w", err)
			}

			fmt.Printf("✓ 已删除 %d 个密钥\n", len(secrets))
			return nil
		},
	}
	secretsCmd.AddCommand(clearCmd)

	rootCmd.AddCommand(secretsCmd)
}

// ── Helper Functions ───────────────────────────────────────────────────────

// loadSecrets loads secrets from the secrets file
func loadSecrets() (map[string]string, error) {
	dataDir, err := getDefaultDataDir()
	if err != nil {
		return nil, err
	}

	secretsPath := filepath.Join(dataDir, "secrets.json")

	// If file doesn't exist, return empty map
	if _, err := os.Stat(secretsPath); os.IsNotExist(err) {
		return make(map[string]string), nil
	}

	data, err := os.ReadFile(secretsPath)
	if err != nil {
		return nil, err
	}

	var secrets map[string]string
	if err := json.Unmarshal(data, &secrets); err != nil {
		return nil, err
	}

	if secrets == nil {
		secrets = make(map[string]string)
	}

	return secrets, nil
}

// saveSecrets saves secrets to the secrets file
func saveSecrets(secrets map[string]string) error {
	dataDir, err := getDefaultDataDir()
	if err != nil {
		return err
	}

	// Ensure data directory exists
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return err
	}

	secretsPath := filepath.Join(dataDir, "secrets.json")

	data, err := json.MarshalIndent(secrets, "", "  ")
	if err != nil {
		return err
	}

	// Save with restrictive permissions (owner read/write only)
	return os.WriteFile(secretsPath, data, 0600)
}

// addToGitignore adds an entry to .gitignore if it doesn't exist
func addToGitignore(gitignorePath, entry string) error {
	// Check if .gitignore exists
	if _, err := os.Stat(gitignorePath); os.IsNotExist(err) {
		// Create .gitignore
		return os.WriteFile(gitignorePath, []byte(entry+"\n"), 0644)
	}

	// Read existing content
	content, err := os.ReadFile(gitignorePath)
	if err != nil {
		return err
	}

	// Check if entry already exists
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == entry {
			return nil // Already exists
		}
	}

	// Append entry
	content = append(content, []byte("\n"+entry+"\n")...)
	return os.WriteFile(gitignorePath, content, 0644)
}
