package main

import (
	"auto-deployment/svc/deploy"
	"fmt"
)

var file = "./playbook.yml"

/*
	var shellRun shell.ShellRun = shell.ShellRun{
		WorkDir: "/tmp",
	}
*/
type SshLoginHandle struct {
	Name string
}

func main() {
	text, err := deploy.Config(file)
	if err != nil {
		fmt.Print(err)
	}
	for _, job := range text.Jobs {
		//fmt.Println("===========", i.JobName, "===========")

		err := job.Validate()
		if err != nil {
			fmt.Println(err)
		} else {
			/*
				fmt.Println("host:  ", i.Hosts)
				fmt.Println("user:  ", i.User)
				fmt.Println("password:  ", i.Password)
				fmt.Println("gorouting:  ", i.ParallelNum)
				fmt.Println("srcFile:  ", i.SrcFile)
				fmt.Println("destDir:  ", i.DestDir)
				fmt.Println("cmd:  ", i.Cmd)
				fmt.Println("shell:  ", i.Shell)
			*/
			//执行任务
			job.RunTask()
		}
	}

}
