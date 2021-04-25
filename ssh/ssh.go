// copy by https://studygolang.com/articles/17990

package ssh

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	gossh "golang.org/x/crypto/ssh"
)

type Cli struct {
	User       string
	Pwd        string
	Addr       string
	Client     *gossh.Client
	Session    *gossh.Session
	LastResult string
}

func (c *Cli) Connect() (*Cli, error) {
	config := &gossh.ClientConfig{}
	config.SetDefaults()
	config.User = c.User
	config.Auth = []gossh.AuthMethod{gossh.Password(c.Pwd)}
	config.HostKeyCallback = func(hostname string, remote net.Addr, key gossh.PublicKey) error { return nil }
	client, err := gossh.Dial("tcp", c.Addr, config)
	if nil != err {
		return c, err
	}
	c.Client = client
	return c, nil
}

// 运行附带环境变量
func (c Cli) RunAsEnv(shell string, kv map[string]string) (string, error) {
	if c.Client == nil {
		if _, err := c.Connect(); err != nil {
			return "", err
		}
	}
	session, err := c.Client.NewSession()
	if err != nil {
		return "", err
	}

	for k, v := range kv {
		fmt.Printf("k: %v, v: %v\n", k, v)
		session.Setenv(k, v)
	}

	defer session.Close()
	buf, err := session.CombinedOutput(shell)

	c.LastResult = string(buf)
	return c.LastResult, err
}

func (c Cli) Run(shell string) (string, error) {
	if c.Client == nil {
		if _, err := c.Connect(); err != nil {
			return "", err
		}
	}
	session, err := c.Client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()
	buf, err := session.CombinedOutput(shell)

	c.LastResult = string(buf)
	return c.LastResult, err
}

func SftpConnect(user, password, host string, port int) (sftpClient *sftp.Client, err error) { //参数: 远程服务器用户名, 密码, ip, 端口
	auth := make([]ssh.AuthMethod, 0)
	auth = append(auth, ssh.Password(password))

	clientConfig := &ssh.ClientConfig{
		User:    user,
		Auth:    auth,
		Timeout: 30 * time.Second,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}

	addr := host + ":" + strconv.Itoa(port)
	sshClient, err := ssh.Dial("tcp", addr, clientConfig) //连接ssh
	if err != nil {
		fmt.Println("连接ssh失败", err)
		return
	}

	if sftpClient, err = sftp.NewClient(sshClient); err != nil { //创建客户端
		fmt.Println("创建客户端失败", err)
		return
	}

	return
}

// Upload file to sftp server
func UploadFile(sc *sftp.Client, srcFile []byte, remoteFile string) (err error) {

	// Make remote directories recursion
	parent := filepath.Dir(remoteFile)
	path := string(filepath.Separator)
	dirs := strings.Split(parent, path)
	for _, dir := range dirs {
		path = filepath.Join(path, dir)
		sc.Mkdir(path)
	}

	// Note: SFTP To Go doesn't support O_RDWR mode
	dstFile, err := sc.OpenFile(remoteFile, (os.O_WRONLY | os.O_CREATE | os.O_TRUNC))
	if err != nil {
		fmt.Fprintf(os.Stderr, "无法打开远程文件: %v\n", err)
		return
	}
	defer dstFile.Close()

	r := bytes.NewReader(srcFile)

	bytes, err := io.Copy(dstFile, r)
	if err != nil {
		fmt.Fprintf(os.Stderr, "无法打开本地文件: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stdout, "%d bytes copied\n", bytes)

	return
}
