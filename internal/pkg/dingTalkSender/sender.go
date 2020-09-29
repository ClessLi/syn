package dingTalkSender

import (
	"fmt"
	"github.com/CatchZeng/dingtalk"
	"strings"
	"time"
)

func sendDingTalkNotification(records map[string]int) {
	var label string
	recordsCount := 0
	if DTSenderConfig.HostInfo.Label == "" {
		label = DTSenderConfig.HostInfo.IP
	} else {
		label = fmt.Sprintf("%s(%s)", DTSenderConfig.HostInfo.Label, DTSenderConfig.HostInfo.IP)
	}
	dateTime := time.Now().Format("2006-01-02 15:04:05.000")
	context := fmt.Sprintf(`[%s] %s环境%s，疑似被暴力登录：`, dateTime, strings.ToUpper(DTSenderConfig.HostInfo.Env), label)
	for s, i := range records {
		recordsCount++
		context = fmt.Sprintf("%s\n\t\t对端主机[%s]已登录失败%d次", context, s, i)
	}

	if recordsCount > 0 {
		//url := fmt.Sprintf(`%s://%s:%d/robot/send?access_token=%s`,
		//	strings.ToLower(DTSenderConfig.DingTalkAPI.Protocol),
		//	DTSenderConfig.DingTalkAPI.Hostname,
		//	DTSenderConfig.DingTalkAPI.Port, DTSenderConfig.DingTalkAPI.Token)

		//fmt.Println(url)
		//fmt.Println(context)

		client := dingtalk.NewClient(DTSenderConfig.DingTalkAPI.Token, DTSenderConfig.DingTalkAPI.Secret)
		msg := dingtalk.NewTextMessage().SetContent(context)
		resp, respErr := client.Send(msg)
		if respErr != nil {
			myLogger.ErrorF("DingTalk msg sending failed: %s", respErr.Error())
		} else {
			myLogger.InfoF("DingTalk response, code: %d, msg: %s", resp.ErrCode, resp.ErrMsg)
		}
	}
}
