package main

import (
	"bufio"
	"context"
	"fmt"
	"github.com/go-vgo/robotgo"
	hook "github.com/robotn/gohook"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	intervalMark = "延迟" //间隔标识
	positionMark = "坐标" //坐标标识
	configDir    = "config"
	execMode     = "exec"
	gatherMode   = "gather"
)

// App struct
type App struct {
	ctx         context.Context
	exitRun     bool
	configName  string
	minInterval int64
	cycle       int
}

type RunParam struct {
	Mode        string `json:"mode"`
	ConfigName  string `json:"configName"`
	MinInterval int64  `json:"minInterval"`
	Cycle       int    `json:"cycle"`
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		exitRun:     false,
		configName:  "mouse.txt",
		minInterval: 500,
	}
}

func (a *App) OnDomReady(ctx context.Context) {
	a.ctx = ctx
	type ModeEnum struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}

	var data struct {
		ConfigName     string     `json:"configName"`
		ConfigFileList []string   `json:"configFileList"`
		ModeEnumList   []ModeEnum `json:"modeEnumList"`
		MinInterval    int64      `json:"minInterval"`
	}
	data.ConfigName = a.configName
	data.ConfigFileList = make([]string, 0)
	data.MinInterval = a.minInterval

	dir, err := os.Open(configDir)
	if err != nil {
		//创建文件
		err := os.MkdirAll(configDir, os.ModePerm)
		if err != nil {
			a.sendErrorMsg("创建配置文件夹失败")
			return
		}
		dir, err = os.Open(configDir)
	}
	defer dir.Close()

	configs, err := dir.Readdir(-1)
	if err != nil {
		a.sendErrorMsg("读取配置文件夹失败")
		return
	}

	data.ModeEnumList = []ModeEnum{
		{Key: gatherMode, Value: "采集模式"},
		{Key: execMode, Value: "执行模式"},
	}

	for _, config := range configs {
		if config.IsDir() {
			continue
		}
		data.ConfigFileList = append(data.ConfigFileList, config.Name())
	}

	runtime.EventsEmit(ctx, "init", data)
}

func (a *App) Run(param RunParam) (res bool) {
	fmt.Println(param)
	hook.Register(hook.KeyDown, []string{"q", "Q"}, func(e hook.Event) {
		a.exitRun = true
		runtime.WindowShow(a.ctx)
		a.sendAlertMsg("程序已退出运行")
		hook.End()
	})
	a.exitRun = false

	if param.ConfigName != "" {
		a.configName = param.ConfigName
	}
	a.minInterval = param.MinInterval
	a.cycle = param.Cycle

	if param.Mode == execMode {
		go a.exec()
	} else {
		a.gatherMousePosition()
	}

	s := hook.Start()
	<-hook.Process(s)
	return true
}

func (a *App) exec() {
	a.runBefore(execMode)
	xArr, yArr, intervalArr := a.getNeedMoveMousePosition()
	cycleCount := 0
	intervalArrLen := len(intervalArr)
	for {
		cycleCount++
		if a.cycle != 0 && cycleCount > a.cycle {
			robotgo.KeyTap("q")
			goto end
		}

		for i := 0; i < len(xArr); i++ {
			if a.exitRun {
				goto end
			}
			robotgo.MoveSmooth(xArr[i], yArr[i], 0.5, 0.5)
			robotgo.Click("left", true)
			wait := int(a.minInterval)
			if i < intervalArrLen {
				wait = intervalArr[i]
			}
			a.sendAlertMsg(fmt.Sprintf("运行中：执行周期%d，【X:%d，Y:%d】 %d毫秒后执行下一次，按q退出", cycleCount, xArr[i], yArr[i], wait))
			time.Sleep(time.Duration(wait) * time.Millisecond)
		}
	}
end:
}

func (a *App) gatherMousePosition() {
	//判断文件是否存在，如果存在，则提示用户
	if _, err := os.Stat(a.getConfigPath()); !os.IsNotExist(err) {
		dialogOptions := runtime.MessageDialogOptions{
			Type:    runtime.QuestionDialog,
			Title:   "提示",
			Message: "配置文件已存在，是否覆盖？",
			Buttons: []string{"是", "否"},
		}
		dialog, err := runtime.MessageDialog(a.ctx, dialogOptions)
		if err != nil {
			runtime.LogError(a.ctx, "弹窗失败")
			return
		}
		runtime.LogInfo(a.ctx, dialog)
		if dialog == "No" {
			a.sendAlertMsg("采集已取消")
			return
		} else {
			//删除文件
			err := os.Remove(a.getConfigPath())
			if err != nil {
				runtime.LogError(a.ctx, "删除配置文件失败")
				return
			}
		}
	}

	a.runBefore(gatherMode)
	a.sendAlertMsg("采集运行中：请点击鼠标左键，按q退出")
	var lastClickTime int64 = 0
	var interval int64 = 0
	hook.Register(hook.MouseDown, []string{}, func(e hook.Event) {
		//如果是左键
		if e.Button == hook.MouseMap["left"] {
			if lastClickTime != 0 {
				//获取两次点击时间间隔(毫秒)
				interval = time.Now().UnixMilli() - lastClickTime
				fmt.Println(interval, a.minInterval)
				//如果间隔小于1秒，不记录
				if interval < a.minInterval {
					interval = a.minInterval
				}
				a.writeConfig(fmt.Sprintf("%s:%d\n", intervalMark, interval))
			}
			lastClickTime = time.Now().UnixMilli()
			a.writeConfig(fmt.Sprintf("%s:%d:%d\n", positionMark, e.X, e.Y))
			a.sendAlertMsg(fmt.Sprintf("采集中：点击x坐标:%d，y坐标:%d，间隔:%d毫秒，按q退出", e.X, e.Y, interval))
		}
	})
}

func (a *App) runBefore(mode string) {
	delay := 3
	title := "采集"
	if mode == execMode {
		title = "执行"
	}
	//每秒发送一次倒计时
	for i := delay; i > 0; i-- {
		a.sendAlertMsg(fmt.Sprintf("%s程序将在%d秒后开始运行，按q退出，倒计时：%d", title, delay, i))
		time.Sleep(time.Second)
	}
	a.sendAlertMsg(fmt.Sprintf("%s程序运行中，按q退出", title))
}

func (a *App) sendAlertMsg(msg string) {
	runtime.EventsEmit(a.ctx, "alertMsg", msg)
}

func (a *App) getNeedMoveMousePosition() (resX []int, resY []int, interval []int) {
	xArr := make([]int, 0)
	yArr := make([]int, 0)
	intervalArr := make([]int, 0)

	file, err := os.Open(a.getConfigPath())
	if err != nil {
		a.sendErrorMsg("配置文件格式错误，读取配置文件失败")
		return xArr, yArr, intervalArr
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		//已:分割
		arr := strings.Split(line, ":")

		if arr[0] == intervalMark {
			intervalArr = append(intervalArr, a.stringParseInt(arr[1]))
			continue
		}

		x := arr[1]
		y := arr[2]
		//转成int16
		xArr = append(xArr, a.stringParseInt(x))
		yArr = append(yArr, a.stringParseInt(y))
	}
	return xArr, yArr, intervalArr
}

func (a *App) stringParseInt(str string) int {
	intValue, err := strconv.Atoi(str)
	if err != nil {
		a.sendErrorMsg("配置文件格式错误，请重新采集")
		panic("配置文件格式错误，请重新采集：" + str)
	}
	return intValue
}

func (a *App) writeConfig(content string) {
	//获取filePath的目录
	filePath := a.getConfigPath()
	dir := filepath.Dir(filePath)
	//如果目录不存在，创建目录
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err := os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			a.sendErrorMsg("创建目录失败")
			return
		}
	}
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		a.sendErrorMsg("读取配置文件失败")
		return
	}
	defer file.Close()

	_, err = file.WriteString(content)
	if err != nil {
		a.sendErrorMsg("配置文件保存失败")
		return
	}
}

func (a *App) getConfigPath() string {
	return fmt.Sprintf("%s/%s", configDir, a.configName)
}

func (a *App) sendErrorMsg(msg string) {
	dialogOptions := runtime.MessageDialogOptions{
		Type:    runtime.ErrorDialog,
		Message: msg,
	}
	runtime.LogError(a.ctx, msg)
	_, _ = runtime.MessageDialog(a.ctx, dialogOptions)
}
