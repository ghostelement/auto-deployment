package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	// 接受参数
	rootCmd.PersistentFlags().String("version", "", "版本")
}

// 根命令
var rootCmd = &cobra.Command{
	Use:   "adp",
	Short: "this is a simple cli app that automates deploy",
	Long: `This is a simple cli app that automates deploy.
e.g. This is a common way to perform deploy, according to playbook.yml in the current path.
Thank's for your support. Please go to https://github.com/ghostelement/auto-deployment and give a star.`,
	Example: "adp run playbook.yml",
	Version: "0.0.5",
}

// 添加一个名为 "version" 的标志（对应 `-v` 和 `--version`）
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of adp",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("adp version: %s\n", rootCmd.Version)
	},
}

// Execute 将所有子命令添加到root命令并适当设置标志。
// 这由 main.main() 调用。它只需要对 rootCmd 调用一次。
func Execute() {
	// 将 "version" 子命令绑定到 rootCmd 上，使其可以通过 `-v` 或 `--version` 调用
	rootCmd.Flags().BoolP("version", "v", false, "Print version information")
	rootCmd.SetVersionTemplate("")
	rootCmd.AddCommand(versionCmd)
	// 执行根命令，并检查执行过程中是否发生了错误。
	cobra.CheckErr(rootCmd.Execute())
}
