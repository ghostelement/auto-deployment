package cmd

import (
	"auto-deployment/svc/db"
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// adp run
var (
	file   string
	dbName string
	dbCmd  = &cobra.Command{
		Use:   "db",
		Short: "Connetion to databases using default playbook",
		Example: `  adp db mydb
  adp db mydb -f /path/to/playbook.yml`,
		Args: cobra.ExactArgs(1), // 最多接收一个参数
		RunE: func(cmd *cobra.Command, args []string) error {
			// 没有参数传递时，使用默认配置文件
			if len(args) == 0 {
				//检查当前目录是否存在配置文件 playbook.yml
				_, err := os.Stat(defConfigName)
				if err != nil {
					if os.IsNotExist(err) {
						return fmt.Errorf("%s", color.RedString("No have db names,please use db name to connetion"))
					}
				}
			} else {
				dbName = args[0]
			}

			//读取配置文件
			if file == "" {
				file = "playbook.yml"
			}

			err := db.ConnetDb(file, dbName)
			if err != nil {
				fmt.Println(err)
			}

			return nil
		},
	}
)

func init() {
	// 添加命令
	dbCmd.Flags().StringVarP(&file, "file", "f", "", "Read databases file from playbook(default playbook.yml)")
	rootCmd.AddCommand(dbCmd)
}
