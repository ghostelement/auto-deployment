package main

import (
	"auto-deployment/logger"
	"auto-deployment/svc/deploy"
	"fmt"
)

var file = "test/test.yml"

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
			//logger.Error("|Error playbook:", file, "|Error: ", err.Error())
			logger.Error("Playbook: ", file, "|Error: ", err.Error())
		} else {
			//执行任务
			job.RunTask()
		}
	}

}
