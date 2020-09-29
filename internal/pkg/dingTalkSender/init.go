package dingTalkSender

import (
	"flag"
	"fmt"
	"github.com/apsdehal/go-logger"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
)

var (
	// 传入参数
	confPath = flag.String("f", "./configs/SynDingTalkSender.yml", "the SynDingTalkSender `config`uration file path.")
	Signal   = flag.String("s", "", "send `signal` to a master process: stop, restart, status")
	help     = flag.Bool("h", false, "this `help`")
	version  = flag.Bool("v", false, "this `version`")

	// 配置对象
	DTSenderConfig = &Config{}
	// 日志变量
	myLogger *logger.Logger
	// 日志文件
	Logf    *os.File
	Stdoutf *os.File

	// 程序工作目录
	workspace string

	// 进程文件
	pidFilename = "SynSender.pid"
	pidFile     string

	// 协程等待组
	wg sync.WaitGroup

	// 错误变量
	procStatusNotRunning = fmt.Errorf("SynDingTalkSender is not running")
)

const (
	VERSION = "v0.0.1"
)

type Config struct {
	DingTalkAPI *DingTalkAPI `yaml:"DingTalkAPI"`
	Syn         *SynConfig   `yaml:"SynConfig,omitempty"`
	HostInfo    *HostInfo    `yaml:"HostInfo"`
}

type DingTalkAPI struct {
	Token  string `yaml:"Token"`
	Secret string `yaml:"Secret"`
	//Hostname string `yaml:"Hostname"`
	//Port     int    `yaml:"Port"`
	//Protocol string `yaml:"Protocol"`
}

type SynConfig struct {
	SecureLogPath       string `yaml:"secureLogPath,omitempty"`
	AuthFailedThreshold int    `yaml:"authFailedThreshold,omitempty"`
	AlarmDelaySec       int    `yaml:"alarmDelaySec,omitempty"`
	SynLogDir           string `yaml:"SynLogDir,omitempty"`
	SynLogLevel         int    `yaml:"SynLogLevel,omitempty"`
}

type HostInfo struct {
	Env   string `yaml:"Env"`
	IP    string `yaml:"IP"`
	Label string `yaml:"HostLabel,omitempty"`
}

// usage, 重新定义flag.Usage 函数，为bifrost帮助信息提供版本信息及命令行工具传参信息
func usage() {
	_, _ = fmt.Fprintf(os.Stdout, `SynDingTalkSender version: %s
Usage: %s [-hv] [-f filename] [-s signal]

Options:
`, VERSION, os.Args[0])
	flag.PrintDefaults()
}

func init() {
	// 初始化工作目录
	ex, pwdErr := os.Executable()
	//fmt.Println(ex)
	if pwdErr != nil {
		panic(pwdErr)
	}

	workspace = filepath.Dir(ex)
	cdErr := os.Chdir(workspace)
	if cdErr != nil {
		panic(cdErr)
	}

	// 初始化pid文件路径
	pidFile = filepath.Join(workspace, pidFilename)

	// 初始化应用传参
	flag.Usage = usage
	flag.Parse()
	if *confPath == "" {
		*confPath = "./configs/SynDingTalkSender.yml"
	}

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	if *version {
		fmt.Printf("SynDingTalkSender version: %s\n", VERSION)
		os.Exit(0)
	}

	// 判断传入配置文件目录
	//*confPath = "F:\\GO_Project\\src\\syn\\cmd\\SynDingTalkSender\\configs\\SynDingTalkSender.yml"
	isExistConfig, pathErr := PathExists(*confPath)
	/* 调测用
	confPath := "./configs/bifrost.yml"
	isExistConfig, pathErr := PathExists(confPath)
	*/
	if !isExistConfig {
		if pathErr != nil {
			fmt.Println("The SynDingTalkSender config file", "'"+*confPath+"'", "is not found.")
		} else {
			fmt.Println("Unkown error of the SynDingTalkSender config file.")
		}
		flag.Usage()
		os.Exit(1)
	}

	// 判断传入信号
	if *Signal != "" && *Signal != "stop" && *Signal != "restart" && *Signal != "status" {
		flag.Usage()
		os.Exit(1)
	}

	// 初始化SynDingTalkSender配置
	confData, readErr := readFile(*confPath)
	//confData, readErr := readFile(confPath)
	if readErr != nil {
		fmt.Println(readErr)
		//myLogger.Error(readErr.Error())
		flag.Usage()
		os.Exit(1)
	}

	yamlErr := yaml.Unmarshal(confData, DTSenderConfig)
	if yamlErr != nil {
		fmt.Println(yamlErr)
		flag.Usage()
		os.Exit(1)
	}
	if DTSenderConfig.Syn == nil {
		DTSenderConfig.Syn = &SynConfig{
			SecureLogPath:       "/var/log/secure",
			AuthFailedThreshold: 5,
			AlarmDelaySec:       600,
			SynLogDir:           "logs",
			SynLogLevel:         3,
		}
	} else {
		if DTSenderConfig.Syn.SecureLogPath == "" {
			DTSenderConfig.Syn.SecureLogPath = "/var/log/secure"
		}
		if DTSenderConfig.Syn.AuthFailedThreshold <= 0 {
			DTSenderConfig.Syn.AuthFailedThreshold = 5
		}
		if DTSenderConfig.Syn.AlarmDelaySec <= 0 {
			DTSenderConfig.Syn.AlarmDelaySec = 600
		}
		if DTSenderConfig.Syn.SynLogDir == "" {
			DTSenderConfig.Syn.SynLogDir = "logs"
		}
		if DTSenderConfig.Syn.SynLogLevel <= 0 {
			DTSenderConfig.Syn.SynLogLevel = 3
		}
	}

	if DTSenderConfig.HostInfo.Label == "" {

	}

	// 初始化日志
	logDir, absErr := filepath.Abs(DTSenderConfig.Syn.SynLogDir)
	if absErr != nil {
		panic(absErr)
	}

	//logDir = "F:\\GO_Project\\src\\syn\\cmd\\SynDingTalkSender\\logs"
	logPath := filepath.Join(logDir, "Syn.log")
	Logf, openErr := os.OpenFile(logPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if openErr != nil {
		panic(openErr)
	}

	var logErr error
	myLogger, logErr = logger.New("Syn", DTSenderConfig.Syn.SynLogLevel, Logf)
	if logErr != nil {
		panic(logErr)
	}
	myLogger.SetFormat("[%{module}] %{time:2006-01-02 15:04:05.000} [%{level}] %{message}\n")
	//fmt.Println(myLogger.Module)
	//myLogger.NoticeF("test")

	// 初始化应用运行日志输出
	stdoutPath := filepath.Join(logDir, "Syn.out")
	Stdoutf, openErr = os.OpenFile(stdoutPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if openErr != nil {
		panic(openErr)
	}
}

// readFile, 读取文件函数
// 参数:
//     path: 文件路径字符串
// 返回值:
//     文件数据
//     错误
func readFile(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	fd, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	return fd, nil
}

// PathExists, 判断文件路径是否存在函数
// 返回值:
//     true: 存在; false: 不存在
//     错误
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil || os.IsExist(err) {
		return true, nil
	} else if os.IsNotExist(err) {
		return false, err
	} else {
		return false, nil
	}
}
