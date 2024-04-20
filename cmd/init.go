package cmd

import (
	"auto-deployment/scripts"
	"auto-deployment/svc/deploy"
	"os"
	"path"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// 默认剧本文件
const (
	defConfigName = "playbook.yml"
	url           = "https://github.com/ghostelement/auto-deployment"
)

// adp init
var (
	initCmd = &cobra.Command{
		Use:   "init",
		Short: `Initialize a new deploy configuration file.`,
		Example: `  adp init
  adp init /path/to/app`,
		Args: cobra.MaximumNArgs(1), // 最多接收1个参数
		RunE: func(cmd *cobra.Command, args []string) error {
			var config *deploy.Playbook
			isAllConfig, err := cmd.Flags().GetBool("all")
			if err != nil {
				return err
			}
			if isAllConfig {
				config = deploy.ExampleAllConfig()
			} else {
				config = deploy.ExampleConfig()
			}
			var appPath string
			// 没有参数传递时，使用默认配置文件
			if len(args) == 0 {
				// 获取当前目录
				dir, err := os.Getwd()
				if err != nil {
					return err
				}
				appPath = dir
			} else {
				appPath = args[0]
			}

			if appPath, err = filepath.Abs(appPath); err != nil {
				return err
			}
			//内置脚本目录
			scriptDir := appPath + "/scripts"
			if scriptDir, err = filepath.Abs(scriptDir); err != nil {
				return err
			}
			//生成配置文件
			confyaml, err := yaml.Marshal(config)
			if err != nil {
				return err
			}

			//写入配置文件
			if err := os.WriteFile(path.Join(appPath, defConfigName), []byte(confyaml), 0644); err != nil {
				return err
			}

			//创建内置脚本目录
			if err := os.MkdirAll(scriptDir, 0755); err != nil {
				return err
			}
			//创建临时脚本目录
			if err := os.MkdirAll(deploy.TmpShellDir, 0755); err != nil {
				return err
			}
			//写入脚本文件
			if err := scripts.CopyShellScriptToWorkingDir(scriptDir); err != nil {
				return err
			}
			return nil
		},
	}
)

func init() {
	// 添加命令
	initCmd.Flags().BoolP("all", "a", false, "init all config")
	rootCmd.AddCommand(initCmd)
}
