package deploy

import (
	"auto-deployment/svc/shell"
	"crypto/rand"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/fatih/color"
)

const (
	line     = "------------------------------------------------------------------------"
	longLine = "====================< %s >===================="
	sortLine = "--- %s ---"
)

var shellRun shell.ShellRun = shell.ShellRun{
	WorkDir: "/tmp",
}

// 本地临时脚本目录
var TmpShellDir = "/tmp/autodeployment/script"

// 远程服务器临时脚本目录
var remoteTmpShellDir = "/tmp/autodeployment/script"

// 以任务名、时间、uuid生成任务id标识
func CreateTaskID(jobname string) string {
	//生成4位数uuid
	b := make([]byte, 2)
	_, erruid := rand.Read(b)
	if erruid != nil {
		fmt.Println("Can't generate random uid")
	}
	uid := fmt.Sprintf("%04x", b)
	return fmt.Sprintf("%s_%s_%s", jobname, time.Now().Format("20060102150405"), uid)
}

// 编写临时脚本，方便远程执行复杂shell
func (job *Job) TmpShell(uuid string, dir string) (string, error) {
	var err error
	// 检查并创建目标目录（如果不存在）
	if _, err = os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			return "", fmt.Errorf("failed to create directory %s: %v", dir, err)
		}
	}
	//临时脚本名
	shellName := fmt.Sprintf("%s.sh", uuid)
	//临时脚本路径
	shell := fmt.Sprintf("%s/%s", dir, shellName)
	//写入临时脚本
	err = os.WriteFile(shell, []byte("#!/bin/bash\n"+job.Shell), 0755)
	if err != nil {
		fmt.Println("shell script write error: ", err)
	}
	return shellName, err

}
func (task *Job) RunTask() {
	var wg sync.WaitGroup
	var outputLock sync.Mutex
	var timemutex sync.Mutex
	var tmpShell string
	var errScp error
	var errShell error
	var errCmd error
	var err error
	// 初始task任务的spinner
	taskSpinner := NewStepSpinner(&outputLock)

	// 记录每个服务器发布&&部署消耗的时间
	var (
		deploytimes  = make(map[string]time.Duration) // {"addr": time}
		deploystatus = make(map[string]string)        // {"addr": SUCCESS}
		sumtime      time.Duration
	)
	//创建5个channal控制并发数量
	if task.ParallelNum == 0 {
		task.ParallelNum = 5
	}
	// Create a channel of type string with a buffer size of task.ParallelNum
	jobChan := make(chan string, task.ParallelNum)
	TaskID := CreateTaskID(task.JobName)

	// 创建临时脚本
	if task.Shell != "" {
		tmpShell, err = task.TmpShell(TaskID, TmpShellDir)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("temp shell: ", tmpShell)
	}

	InfoF(longLine, "Task: "+task.JobName)
	fmt.Println("Task ID: ", TaskID)
	//输出task执行动画
	taskSpinner.Start(task.JobName)
	defer taskSpinner.Stop()
	for _, host := range task.Hosts {
		wg.Add(1)

		go func(host string) {
			defer wg.Done()
			jobChan <- host
			startime := time.Now() // 记录开始时间
			sshArgs := shell.SshLoginArgs{
				Host:     host,
				User:     task.User,
				Password: task.Password,
			}
			//scp文件到远程服务器指定目录
			if task.SrcFile != "" {
				if task.DestDir == "" {
					task.DestDir = remoteTmpShellDir
				}
				//scpSpinner.Start(fmt.Sprint("[", host, "]", " SCP"))
				errScp = shellRun.Scp(sshArgs, task.SrcFile, task.DestDir)
				//scpSpinner.Stop()
			}
			//执行cmd命令
			if task.Cmd != "" {
				//cmdSpinner.Start(fmt.Sprint("[", host, "]", " CMD"))
				errCmd = shellRun.SshLoginAndRun(sshArgs, task.Cmd, []string{"", ""}, func(name, msg string) {
					fmt.Printf("\n[[HOST CMD]]>>[%s]:\n%s\n", name, msg)
				})
				//cmdSpinner.Stop()
			}
			//执行shell命令
			if tmpShell != "" {
				//shellSpinner.Start(fmt.Sprint("[", host, "]", " SHELL"))
				// scp临时脚本到目标服务器
				shellRun.Scp(sshArgs, (TmpShellDir + "/" + tmpShell), remoteTmpShellDir)
				//shellRun.Scp(sshArgs, TmpShellDir, remoteTmpShellDir)
				//切换远程临时脚本目录并执行临时脚本
				errShell = shellRun.SshLoginAndRun(sshArgs, "cd "+remoteTmpShellDir+";bash", []string{"", tmpShell}, func(name, msg string) {
					fmt.Printf("\n[[HOST SHELL]]>>[%s]:\n%s\n", name, msg)
				})
				//shellSpinner.Stop()
			}
			//计算任务耗时,用互斥锁防止多个进程同时写入
			timemutex.Lock()
			deploytimes[host] = time.Since(startime)
			sumtime += deploytimes[host]
			timemutex.Unlock()
			if err == nil && errShell == nil && errCmd == nil && errScp == nil {
				timemutex.Lock()
				deploystatus[host] = "SUCCESS"
				timemutex.Unlock()
			} else {
				timemutex.Lock()
				deploystatus[host] = "FAILED"
				timemutex.Unlock()
			}
			<-jobChan
		}(host)
	}
	wg.Wait()
	close(jobChan)
	// 打印汇总信息
	printSummary(deploytimes, deploystatus, sumtime)
}

// printSummary prints the summary of the deployment
func printSummary(deploytimes map[string]time.Duration, deploystatus map[string]string, sumtime time.Duration) {
	Info(line)
	Info("Summary:")
	Info("")
	for addr, deploytime := range deploytimes {
		//总长度固定52
		//计算需要补充的.的个数
		dotNum := 52 - len(addr)
		dotStr := ""
		for i := 0; i < dotNum; i++ {
			dotStr += "."
		}
		//执行成功还是失败设置颜色
		var statuscolor string
		if deploystatus[addr] == "SUCCESS" {
			statuscolor = color.GreenString(deploystatus[addr])
		} else {
			statuscolor = color.RedString(deploystatus[addr])
		}
		InfoF("%s %s %s [  %f s]", addr, dotStr, statuscolor, deploytime.Seconds())
	}
	Info(line)
	Info(color.GreenString("TASK END"))
	Info(line)
	InfoF("Total time: %f s", sumtime.Seconds())
	InfoF("Finished at: %s", time.Now().Format("2006-01-02 15:04:05"))
	Info(line)
}
