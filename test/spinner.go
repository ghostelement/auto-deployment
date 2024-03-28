package main

import (
	"fmt"
	"sync"
	"time"
)

var wg sync.WaitGroup
var outputLock sync.Mutex

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
				fmt.Printf("\r%s: Running... [%s]", stepName, frames[i%len(frames)])
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

func main() {
	hosts := []string{"11111", "22222", "33333", "44444", "55555", "66666", "77777", "88888", "99999"}

	// 初始化各个步骤的spinner
	scpSpinner := NewStepSpinner(&outputLock)
	jobChan := make(chan string, 5)
	hostID := 0
	scpSpinner.Start(fmt.Sprint("[The task]", ""))
	defer scpSpinner.Stop()
	for _, host := range hosts {
		wg.Add(1)
		hostID++
		go func(host string, hostID int) {
			defer wg.Done()
			jobChan <- host
			scprun(hostID)
			<-jobChan
		}(host, hostID)
	}
	wg.Wait()
	close(jobChan)

}

func scprun(hostID int) {
	fmt.Println("\n SCPing host:", hostID)
	time.Sleep(time.Second * 10)
}
