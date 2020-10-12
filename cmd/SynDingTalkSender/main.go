package main

import (
	"errors"
	"fmt"
	"github.com/ClessLi/syn/internal/pkg/dingTalkSender"
	"os"
	"time"
)

func main() {
	defer dingTalkSender.Logf.Close()
	defer dingTalkSender.Stdoutf.Close()

	err := errors.New("unkown signal")
	//err = dingTalkSender.Start()
	//if err == nil {
	//	fmt.Println("SynDingTalkSender is stopped")
	//	os.Exit(0)
	//}
	switch *dingTalkSender.Signal {
	case "":
		err = dingTalkSender.Start()
		if err == nil {
			fmt.Printf("%s SynDingTalkSender is stopped.\n", time.Now().Format("2006-01-02 15:04:05.000"))
			os.Exit(0)
		}
	case "stop":
		err = dingTalkSender.Stop()
		if err == nil {
			fmt.Printf("%s SynDingTalkSender is finished.\n", time.Now().Format("2006-01-02 15:04:05.000"))
			os.Exit(0)
		}
	case "restart":
		err = dingTalkSender.Restart()
		if err == nil {
			fmt.Printf("%s SynDingTalkSender was restarted, and it's stopped, now.\n", time.Now().Format("2006-01-02 15:04:05.000"))
			os.Exit(0)
		} else if err == dingTalkSender.SubProcessStarted {
			fmt.Printf("%s SynDingTalkSender was restarted, and is running, now.\n", time.Now().Format("2006-01-02 15:04:05.000"))
			os.Exit(0)
		}
	case "status":
		pid, statErr := dingTalkSender.Status()
		if statErr != nil {
			fmt.Printf("SynDingTalkSender is abnormal with error: %s\n", statErr.Error())
			os.Exit(1)
		} else {
			fmt.Printf("SynDingTalkSender <PID %d> is running\n", pid)
			os.Exit(0)
		}
	}
	if err == dingTalkSender.SubProcessStarted {
		fmt.Printf("%s SynDingTalkSender is started.\n", time.Now().Format("2006-01-02 15:04:05.000"))
		os.Exit(0)
	}
	fmt.Println(err.Error())
	os.Exit(1)
}
