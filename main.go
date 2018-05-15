package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"
	"fmt"
	"runtime"
	"os/exec"
	"path/filepath"
	"path"
	"log"
	"net/http"
	"strings"
)

type PairItem []string //[first,second]

type FileConfig struct {
	DirMap []PairItem `json:"dirmap"`
}

/*
* 用法：
fileserver.json
{
  "dirmap":[
    ["/","../"],
    ["/t1/","wow/next/"],
    ["/imgs/2018/5/15/","resources/images/20180515"]
  ]
}
其中dirmap数组子项，如["/t1/","wow/next/"]中"/t1/"对应的url中的path，"wow/next"对应实际相对可执行目录的资源目录
*/

const ConfigFileName = "fileserver.json"

func main() {
	log.Println("Starting FileServer...")
	log.Println("OS:", OperateSystem(), "Arch:", SysArch())
	go listenOSSignal()
	fileConfig := loadConfigFile()
	for _, item := range fileConfig.DirMap {
		log.Println(item[0], "->", item[1])
		netPath := item[0]
		if !strings.HasPrefix(netPath, "/") {
			netPath = "/" + netPath
		}
		if !strings.HasSuffix(netPath, "/") { //如果目录后没有/则追加/
			netPath += "/"
		}
		http.Handle(netPath, http.StripPrefix(netPath, http.FileServer(http.Dir(item[1]))))
	}
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}

func loadConfigFile() (fileConf FileConfig) {
	data, err := ioutil.ReadFile(ConfigFileName)
	if err != nil {
		panic(err.Error())
		return
	}
	err = json.Unmarshal(data, &fileConf)
	if err != nil {
		panic(err.Error())
		return
	}
	return fileConf
}

//重新加载配置文件
func reloadConfigFile() {

}

func listenOSSignal() {
	c := make(chan os.Signal)
	signal.Notify(c)
	for s := range c {
		log.Println("接受到信号：", s)
		if (OperateSystem() == "windows") {
			SignalHandle(s)
		}
	}
}

//FIXME: 由外部进程发命令信号并对其解析
//命令：fileServer [-reload] -config fileserver.json
func SignalHandle(s os.Signal) {
	switch s {
	case os.Interrupt:
		os.Exit(1)
	case syscall.SIGINT:
		fmt.Println(">>", s)
	}
}

func SysArch() string {
	return runtime.GOARCH
}

func OperateSystem() string {
	return runtime.GOOS
}

//可执行程序路径
func ExePath() string {
	file, _ := exec.LookPath(os.Args[0])
	exepath, _ := filepath.Abs(file)
	return exepath
}

//当前源代码路径
func WorkPath() string {
	_, filename, _, _ := runtime.Caller(1)
	return filename
}

//源代码工作目录
func WorkDir() string {
	return path.Dir(WorkPath())
}

//返回参数filename在源代码共目录的绝对路径
func AbsPathInWorkDir(filename string) string {
	return path.Join(WorkDir(), filename)
}
