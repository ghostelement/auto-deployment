package main

import (
	"auto-deployment/svc/deploy"
	"embed"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/fatih/color"

	"github.com/urfave/cli/v2"
)

//go:embed scripts/*
var scripts embed.FS

//引用scripts目录内脚本文件

const (
	defConfigName = "playbook.yml"
	url           = "https://gitee.com/ghostelement/auto-deployment"
)

var (
	Version = "v0.0.1"
	Os      = "linux"
	Arch    = "amd64"
)

func main() {
	cli.VersionPrinter = func(ctx *cli.Context) {
		fmt.Printf("deploy version %s %s/%s\r\n", ctx.App.Version, Os, Arch)
	}

	app := &cli.App{
		Name: "adp",
		Description: `This is a simple cli app that automates deploy.
e.g. This is a common way to perform deploy, according to dyplaybook.yml in the current path
	adp
This is manually specifying the configuration file
	adp /path/to/playbook.yml`,
		Usage:     "this is a simple cli app that automates deploy",
		UsageText: `adp [/path/to/playbook.yml]`,
		Version:   Version,
		Action: func(ctx *cli.Context) error {
			// 网址
			deploy.Info(color.BlueString("Thank you for your support. You can go to %s and give a star.", url))
			cli.ShowAppHelp(ctx)
			return nil
		},
		// adp命令行
		Commands: []*cli.Command{
			{
				//adp run
				Name: "run",
				Description: `Run auto deployment from playbook
		EXM: adp run
Run auto deployment from your playbook
		EXM: adp run /path/to/playbook.yml
`,
				UsageText: `adp run [/path/to/playbook.yml]`,
				Action: func(ctx *cli.Context) (err error) {
					profile := ctx.Args().First()
					if profile == "" {
						//检查当前目录是否存在配置文件 playbook.yml
						_, err := os.Stat(defConfigName)
						if err != nil {
							if os.IsNotExist(err) {
								return fmt.Errorf("playbook.yml does not exist, please use 'deploy init' to initialize")
							}
							return err
						}
						profile = defConfigName
					}

					//读取配置文件
					config, err := deploy.Config(profile)
					if err != nil {
						return err
					}

					for _, job := range config.Jobs {
						err := job.Validate()
						if err != nil {
							fmt.Println(err)
						} else {
							//执行任务
							job.RunTask()
						}
					}
					return nil
				},
			},
			{
				//adp init
				Name: "init",
				Description: `Initialize a new deploy configuration file.
e.g. The usual way to config an app
		adp init
The specified application directory has been initially configured
		adp init /path/to/app
`,
				UsageText: `adp init [/path/to/app]`,
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:        "a",
						DefaultText: "false",
						Usage:       "All configurations",
					},
				},
				Action: func(ctx *cli.Context) (err error) {

					isAllConfig := ctx.Bool("a")

					appath := ctx.Args().First()
					if appath == "" {
						//获取程序所在目录
						exePath, err := os.Executable()
						if err != nil {
							fmt.Println(err)
							return err
						}
						dirPath := filepath.Dir(exePath)
						fmt.Println("The exe path: ", dirPath)
						appath = dirPath
					}

					if appath, err = filepath.Abs(appath); err != nil {
						return err
					}

					//临时目录及脚本目录
					scriptDir := appath + "/scripts"
					tmpDir := appath + "/tmp/script"
					//fmt.Println("appath:" + appath)

					var config *deploy.Playbook
					if isAllConfig {
						config = deploy.ExampleConfig()
					} else {
						config = deploy.ExampleConfig()
					}

					//生成配置文件
					confyaml, err := yaml.Marshal(config)
					if err != nil {
						return err
					}

					//写入配置文件
					dpyconfig, err := os.Create(path.Join(appath, defConfigName))
					if err != nil {
						return err
					}
					//创建目录
					if err := os.MkdirAll(scriptDir, 0755); err != nil {
						return err
					}
					if err := os.MkdirAll(tmpDir, 0755); err != nil {
						return err
					}
					//写入脚本文件
					if err := CopyShellScriptToWorkingDir(scriptDir); err != nil {
						return err
					}
					//goland:noinspection GoUnhandledErrorResult
					defer dpyconfig.Close()
					if _, err := io.WriteString(dpyconfig, string(confyaml)); err != nil {
						return err
					}

					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		deploy.Error(err.Error())
		os.Exit(1)
	}
}

// 拷贝scripts目录内脚本文件
func CopyShellScriptToWorkingDir(destPath string) error {
	files, err := scripts.ReadDir("scripts")
	if err != nil {
		return err
	}

	for _, file := range files {
		filename := file.Name()
		fmt.Println("Write scripts: ", filename)
		data, err := scripts.ReadFile("scripts/" + filename)
		if err != nil {
			return err
		}
		err = os.WriteFile(filepath.Join(destPath, filename), data, 0755)
		if err != nil {
			return err
		}
	}

	return err
}
