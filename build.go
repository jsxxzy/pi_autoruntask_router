// [该脚本用于自动上传脚本]

package main

import (
	_ "embed"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/d1y/router/ssh"
)

//go:embed index.html
var htmlCode string

//go:embed inetd
var daemonBin []byte

//go:embed daemon.sh
var daemonScript []byte

func headers(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, htmlCode)
}

// 小米路由器网关
var xiaomiRouterRPC = fmt.Sprintf("%v:%v", "10.32.0.1", "22")

var xiaomiRouterSSHAuth = "admin"

func routerRunSSHCmd(cmd string) (string, error) {

	cli := ssh.Cli{
		User: xiaomiRouterSSHAuth,
		Pwd:  xiaomiRouterSSHAuth,
		Addr: xiaomiRouterRPC,
	}

	return cli.Run(cmd)

}

func routerRunSSHCmdAndEnv(cmd string, kv map[string]string) (string, error) {

	cli := ssh.Cli{
		User: xiaomiRouterSSHAuth,
		Pwd:  xiaomiRouterSSHAuth,
		Addr: xiaomiRouterRPC,
	}

	// TODO
	return cli.RunAsEnv(cmd, kv)

}

// 检查后台运行程序是否存在
var checkDaemonFile = `[[ -f /tmp/app/inetd ]] && echo "1"`

// 检查后台程序是否运行
func checkFunc() (bool, string) {
	var output, err = routerRunSSHCmd(checkDaemonFile)

	if err != nil {
		return false, fmt.Sprintf("发送错误: %v", err.Error())
	}

	output = strings.TrimSpace(output)

	if output == "1" {
		return true, "文件存在, 代表后台程序正在运行"
	} else {
		fmt.Println("output: ", output, "len: ", len(output))
		return false, "// 文件不存在, 后台没运行\n// 多种情况, 可能是断电导致的"
	}

}

// 上传后台程序
//
// 包括二进制文件, 脚本文件
func uploadDaemon() error {
	var sftpBox, _ = ssh.SftpConnect("admin", "admin", "10.32.0.1", 22)
	var err1 = ssh.UploadFile(sftpBox, daemonBin, "/tmp/app/inetd")
	var err2 = ssh.UploadFile(sftpBox, daemonScript, "/tmp/app/daemon.sh")
	if err1 != nil || err2 != nil {
		return errors.New(err1.Error() + "\n" + err2.Error())
	}
	return nil
}

// 检测服务
func handleCheckAction(w http.ResponseWriter, req *http.Request) {
	var flag, msg = checkFunc()
	var outMsg = "已运行"
	if !flag {
		outMsg = "未运行"
	}
	fmt.Fprintf(w, "%v\n%v", outMsg, msg)
}

// 开启服务
func handleStartAction(w http.ResponseWriter, req *http.Request) {

	var flag, msg = checkFunc()

	msg += "\n"

	if !flag {
		var err = uploadDaemon()
		if err != nil {
			msg += "上传后台程序失败\n"
		} else {
			var pip, err = routerRunSSHCmd("sh /tmp/app/daemon.sh")
			if err != nil {
				msg += fmt.Sprintf("未知错误: ", err.Error())
			} else {
				msg += fmt.Sprintf("执行成功: ", pip)
			}
			msg += "\n"
		}
	}

	fmt.Fprintf(w, "%v", msg)
}

func main() {
	if len(os.Args) <= 1 {
		panic("请设置端口")
	}
	var port = os.Args[1]
	if port == "" {
		port = "8080"
	}
	var targetPort, err = strconv.Atoi(port)
	if err != nil {
		panic("端口设置错误")
	}
	var formatPort = fmt.Sprintf(":%v", targetPort)
	http.HandleFunc("/", headers)
	http.HandleFunc("/api/check", handleCheckAction)
	http.HandleFunc("/api/start", handleStartAction)
	log.Fatal(http.ListenAndServe(formatPort, nil))
}
