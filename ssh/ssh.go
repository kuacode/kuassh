package ssh

import (
	"fmt"
	"github.com/containerd/console"
	"github.com/kuassh"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"runtime"
	"syscall"
	"time"
)

var (
	DefaultCiphers = []string{
		"aes128-ctr",
		"aes192-ctr",
		"aes256-ctr",
		"aes128-gcm@openssh.com",
		"chacha20-poly1305@openssh.com",
		"arcfour256",
		"arcfour128",
		"arcfour",
		"aes128-cbc",
		"3des-cbc",
		"blowfish-cbc",
		"cast128-cbc",
		"aes192-cbc",
		"aes256-cbc",
	}
)

type client struct {
	Node          *kuassh.Node
	SSHClientConf *ssh.ClientConfig
	SSHClient     *ssh.Client
	session       *ssh.Session
	osName        string
	win           *terminalWindow // 窗口
}

type terminalWindow struct {
	h int // 高
	w int // 宽
}

func NewClient(n *kuassh.Node) (*client, error) {
	auth := make([]ssh.AuthMethod, 0)

	if n.KeyFile != "" {
		keyByte, err := ioutil.ReadFile(n.KeyFile)
		signer, err := ssh.ParsePrivateKey(keyByte)
		if err != nil {
			return nil, err
		}
		auth = append(auth, ssh.PublicKeys(signer))
	}
	if n.PassWord != "" {
		auth = append(auth, ssh.Password(n.PassWord))
	}

	if len(auth) == 0 {
		fmt.Printf("password:")
		b, err := terminal.ReadPassword(int(syscall.Stdin))
		if err != nil {
			return nil, err
		}
		auth = append(auth, ssh.Password(string(b)))
	}
	sshConfig := &ssh.ClientConfig{
		User:            n.User,
		Auth:            auth,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         time.Second * 30,
	}

	sshConfig.SetDefaults()
	sshConfig.Ciphers = append(sshConfig.Ciphers, DefaultCiphers...)

	c := &client{
		SSHClientConf: sshConfig,
		Node:          n,
		osName:        runtime.GOOS,
	}
	return c, nil
}

func (c *client) Login() *ssh.Client {
	var err error
	host := c.Node.Host
	port := c.Node.Port
	//
	jn := c.Node.Jump
	if len(jn) > 0 {
		jnc, err := NewClient(jn[0])
		if err != nil {
			log.Fatal("创建jump节点错误:", err)
		}
		proxyClient, err := ssh.Dial("tcp", net.JoinHostPort(jnc.Node.Host, jnc.Node.Port), jnc.SSHClientConf)
		if err != nil {
			log.Fatal(err)
		}
		conn, err := proxyClient.Dial("tcp", net.JoinHostPort(host, port))
		if err != nil {
			log.Fatal(err)
		}
		ncc, chans, reqs, err := ssh.NewClientConn(conn, net.JoinHostPort(host, port), c.SSHClientConf)
		if err != nil {
			log.Fatal(err)
		}
		c.SSHClient = ssh.NewClient(ncc, chans, reqs)
	} else {
		c.SSHClient, err = ssh.Dial("tcp", net.JoinHostPort(c.Node.Host, c.Node.Port), c.SSHClientConf)
		if err != nil {
			log.Fatal("登陆错误:", err)
		}
	}
	//c.StartSession()
	return c.SSHClient
}

func (c *client) StartSession() {
	defer c.SSHClient.Close()
	//
	var err error
	c.session, err = c.SSHClient.NewSession()
	if err != nil {
		log.Fatal("NewSession:", err)
	}
	defer c.session.Close()
	//
	current := console.Current()
	defer current.Close()
	// 去掉缓存
	err = current.SetRaw()
	if err != nil {
		log.Fatal("MakeRaw:", err)
	}
	// 退出还原终端
	defer current.Reset()
	// 终端大小;windows 下获取输出才能正确运行,目前linux和windows下获取输出调整窗口大小正常，暂时不做区分处理
	// 获取终端大小
	ws, err := current.Size()
	if err != nil {
		log.Fatal("GetSize:", err)
	}
	c.win = &terminalWindow{
		h: int(ws.Height),
		w: int(ws.Width),
	}
	// 监听窗口变化
	go c.winChange(current)
	// 请求Pty
	c.requestPty()
	// 重定向输入输出
	//c.session.Stdout = os.Stdout
	//c.session.Stderr = os.Stderr
	//c.session.Stdin = os.Stdin
	// 直接对接了 stderr、stdout 和 stdin 会造成 tmux等出问题 ，实际上我们应当启动一个异步的管道式复制行为
	stdoutPipe, err := c.session.StdoutPipe()
	if err != nil {
		log.Fatal("StdoutPipe", err)
	}
	go func(r io.Reader) {
		_, _ = io.Copy(os.Stdout, r)
	}(stdoutPipe)
	//
	stderrPipe, err := c.session.StderrPipe()
	if err != nil {
		log.Fatal("StderrPipe", err)
	}
	go func(r io.Reader) {
		_, _ = io.Copy(os.Stderr, r)
	}(stderrPipe)

	stdinPipe, err := c.session.StdinPipe()
	if err != nil {
		log.Fatal("StdinPipe", err)
	}
	// 系统终端输入拷贝到远程终端执行
	go c.runInput(current, stdinPipe)
	//
	c.shell()
	// 初始化命令
	c.runCmds(stdinPipe)
	//
	go c.keepalive()
	// 等待shell
	err = c.session.Wait()
	if err != nil {
		log.Fatal("Wait", err)
	}
}

// 系统终端输入拷贝到远程终端执行
func (c *client) runInput(current console.Console, w io.Writer) {
	buf := make([]byte, 128)
	for {
		n, err := current.Read(buf)
		if err != nil {
			log.Fatal("终端读取命令错误:", err)
		}
		if n > 0 {
			_, err = w.Write(buf[:n])
			if err != nil {
				log.Fatal("发送命令错误:", err)
			}
		}
	}
}

func (c *client) requestPty() {
	modes := ssh.TerminalModes{
		ssh.ECHO:          1,     // 禁用回显（0禁用，1启动）
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, //output speed = 14.4kbaud
	}
	// default to xterm-256color
	termType := os.Getenv("xterm")
	if termType == "" {
		termType = "xterm-256color"
	}
	// request pty
	err := c.session.RequestPty(termType, c.win.h, c.win.w, modes)
	//err = session.RequestPty("xterm", h, w, modes)
	if err != nil {
		log.Fatal("RequestPty", err)
	}
}

func (c *client) runCmds(w io.Writer) {
	// todo 执行初始化命令
	for i := range c.Node.Cmds {
		shellCmd := c.Node.Cmds[i]
		time.Sleep(shellCmd.Delay * time.Millisecond)
		w.Write([]byte(shellCmd.Cmd + "\r"))
	}
}

func (c *client) keepalive() {
	// 每30s发送一次信号
	t := time.Tick(30 * time.Second)
	for range t {
		// 保持连接
		c.session.SendRequest("keepalive", true, nil)
	}
}

func (c *client) shell() {
	// 开启shell
	err := c.session.Shell()
	if err != nil {
		log.Fatal("Shell", err)
	}
}
