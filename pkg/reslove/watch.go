package reslove

import (
	"github.com/apsdehal/go-logger"
	"github.com/hpcloud/tail"
	"os"
	"regexp"
	"time"
)

var (
	SecurePath                = "/var/log/secure"
	hostBruteLoginRecordMatch = regexp.MustCompile(HostBruteLoginRecordReg).FindStringSubmatch
	//hostBruteLoginRecordWithUserMatch = regexp.MustCompile(HostBruteLoginRecordWithUserReg).FindStringSubmatch
	Hosts     = NewHosts()
	Threshold = 5
)

func WatchSecure(log *logger.Logger) {
	tails, tailErr := tail.TailFile(SecurePath, tail.Config{
		//tails, tailErr := tail.TailFile(filepath, tail.Config{
		ReOpen:    true,
		Follow:    true,
		Location:  &tail.SeekInfo{Offset: 0, Whence: os.SEEK_END},
		MustExist: false,
		Poll:      true,
	})
	if tailErr != nil {
		log.ErrorF(tailErr.Error())
		return
	}
	var msg *tail.Line
	var ok bool

	log.Info("Start watching secure log")
	for {
		msg, ok = <-tails.Lines
		if !ok {
			log.NoticeF("tail file close reopen, filename:%s\n", tails.Filename)
			time.Sleep(10 * time.Millisecond)
			continue
		}

		match := hostBruteLoginRecordMatch(msg.Text)
		if len(match) != 2 {
			log.DebugF("match err, msg:", msg.Text)
			continue
		}
		host := match[1]

		parseErr := Hosts.AuthFailed(host, Threshold)
		if parseErr != nil {
			log.NoticeF(parseErr.Error())
		}
	}
}

func WatchAlarm(log *logger.Logger, alarmDelay time.Duration, num int, callback func(map[string]int)) {
	log.Notice("Start monitoring alarm record")
	for {
		records := make(map[string]int)
		for !Hosts.alarm.IsEmpty() {
			ipv4, ok := Hosts.alarm.Poll().(string)
			if ok && Hosts.visits[ipv4].ThresholdReached(num) && !Hosts.visits[ipv4].isAlarmed {
				count := Hosts.visits[ipv4].authFailureCount
				log.WarningF("Host with IP of %s is suspected to be brutally logged in. Access denied %d times", ipv4, count)
				records[ipv4] = count
				Hosts.visits[ipv4].isAlarmed = true
				rfErr := Hosts.ResetHostFailCount(ipv4)
				if rfErr != nil {
					log.ErrorF(rfErr.Error())
				}
			}
		}
		for s, host := range Hosts.visits {
			if time.Now().Sub(host.timestamp) >= alarmDelay {
				if host.ThresholdReached(num) {
					log.WarningF("Host with IP of %s is suspected to be brutally logged in. Access has been denied %d times since %s",
						s,
						host.authFailureCount,
						host.timestamp.Format("2006-01-02 15:04:05.000"))
					records[s] = host.authFailureCount
				}
				rrErr := Hosts.ResetHostRecord(s)
				if rrErr != nil {
					log.ErrorF(rrErr.Error())
				}
			}
		}
		//for host := range Hosts.alarm {
		//	if Hosts.alarm[host] {
		//		count := Hosts.visits[host].authFailureCount
		//		records[host] = count
		//		log.WarningF("IP for the host of %s, suspected violent login. The number of access denied is %d", host, count)
		//		parseErr := Hosts.ResetHostFailCount(host)
		//		if parseErr != nil {
		//			log.ErrorF(parseErr.Error())
		//		}
		//	}
		//}
		//if len(records) > 0 {
		//	callback(records)
		//}
		callback(records)
		//for s, _ := range records {
		//	parseErr := Hosts.ResetHostFailCount(s)
		//	if parseErr != nil {
		//		log.ErrorF(parseErr.Error())
		//	}
		//}
		time.Sleep(time.Millisecond * 10)
	}
}
