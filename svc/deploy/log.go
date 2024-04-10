package deploy

import (
	"fmt"

	"sync"
	"time"

	"github.com/fatih/color"
)

// 前台日志输出格式
func InfoF(format string, args ...any) {
	fmt.Println(fmt.Sprintf(format, args...))
}

func Info(msg string) {
	fmt.Println(msg)
}

func SuccessF(format string, args ...any) {
	fmt.Println(fmt.Sprintf("%s\t\t%s", fmt.Sprintf(format, args...), color.GreenString("[ok]")))
}

func Success(msg string) {
	fmt.Println(fmt.Sprintf("%s\t\t%s", msg, color.GreenString("[ok]")))
}

func WarnF(format string, args ...any) {
	fmt.Println(color.YellowString(format, args...))
}

func Warn(msg string) {
	fmt.Println(color.YellowString(msg))
}

func ErrorF(format string, args ...any) {
	fmt.Println(color.RedString(format, args...))
}

func Error(msg string) {
	fmt.Println(color.RedString(msg))
}

func InfoShell(cmd string) {
	fmt.Printf("%s %s\r\n", color.GreenString("~"), cmd)
}

// 任务执行动画
type StepSpinner struct {
	SpinnerStopChan chan bool
	OutputLock      *sync.Mutex
}

func NewStepSpinner(outputLock *sync.Mutex) *StepSpinner {
	return &StepSpinner{
		SpinnerStopChan: make(chan bool),
		OutputLock:      outputLock,
	}
}

// 任务执行动画格式
func (ss *StepSpinner) Start(stepName string) {
	go func() {
		frames := []string{"-", "\\", "|", "/"}
		i := 0
		for {
			select {
			case <-ss.SpinnerStopChan:
				ss.OutputLock.Lock()
				fmt.Println("") // 清除当前行的spinner
				ss.OutputLock.Unlock()
				return
			default:
				ss.OutputLock.Lock()
				fmt.Printf("\r[Task %s]: Running... [%s]", stepName, frames[i%len(frames)])
				ss.OutputLock.Unlock()
				i++
				time.Sleep(time.Millisecond * 300)
			}
		}
	}()
}

func (ss *StepSpinner) Stop() {
	ss.SpinnerStopChan <- true
}
