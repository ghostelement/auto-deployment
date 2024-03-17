package main

import (
	"auto-deployment/svc/deploy"
	"fmt"
)

func TestYaml(file string) {
	text, err := deploy.Config(file)
	if err != nil {
		fmt.Print(err)
	}
	for _, i := range text.Jobs {
		fmt.Println("===========", i.JobName, "===========")
		fmt.Println("host:  ", i.Hosts)
		fmt.Println("user:  ", i.User)
		fmt.Println("password:  ", i.Password)
		fmt.Println("gorouting:  ", i.ParallelNum)
		fmt.Println("srcFile:  ", i.SrcFile)
		fmt.Println("destDir:  ", i.DestDir)
		fmt.Println("cmd:  ", i.Cmd)
		fmt.Println("shell:  ", i.Shell)
	}
}

func main() {
	TestYaml("./playbook.yml")
}
