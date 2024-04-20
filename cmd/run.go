package cmd

import (
	"auto-deployment/logger"
	"auto-deployment/svc/deploy"
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// adp run
var (
	runCmd = &cobra.Command{
		Use:   "run",
		Short: "Run auto deployment using default playbook",
		Example: `  adp run
  adp run /path/to/playbook.yml`,
		Args: cobra.MaximumNArgs(1), // 最多接收一个参数
		RunE: func(cmd *cobra.Command, args []string) error {
			var profile string
			// 没有参数传递时，使用默认配置文件
			if len(args) == 0 {
				//检查当前目录是否存在配置文件 playbook.yml
				_, err := os.Stat(defConfigName)
				if err != nil {
					if os.IsNotExist(err) {
						return fmt.Errorf(color.RedString("playbook.yml does not exist, please use 'deploy init' to initialize"))
					}
				}
				profile = defConfigName
			} else {
				profile = args[0]
			}

			//读取配置文件
			config, err := deploy.Config(profile)
			if err != nil {
				return err
			}

			//验证job任务字段是否合规并执行任务
			for _, job := range config.Jobs {
				err := job.Validate()
				if err != nil {
					logger.Error("|Error playbook:", profile, "|Error: ", err.Error())
					fmt.Println(color.RedString("Error: "), err)
				} else {
					//执行任务
					job.RunTask()
				}
			}
			return nil
		},
	}
)

func init() {
	// 添加命令
	rootCmd.AddCommand(runCmd)
}
