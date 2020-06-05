// +build !windows

package ssh

import (
	"golang.org/x/crypto/ssh/terminal"
	"log"
	"os"
	"os/signal"
	"syscall"
)

// 监听窗口变化
// 非windows下可以监听信号
// sigwinchCh := make(chan os.Signal, 1)
// signal.Notify(sigwinchCh, syscall.SIGWINCH)
// for {
// 	select {
//		case sigwinchCh:
//			...
//	}
// }
// 监听窗口大小变化
func (c *client) winChange(fd int) {
	sigwinchCh := make(chan os.Signal, 1)
	signal.Notify(sigwinchCh, syscall.SIGWINCH)

	for {
		select {
		case <-sigwinchCh:
			currTermWidth, currTermHeight, err := terminal.GetSize(fd)
			if err != nil {
				log.Printf("获取当前窗口大小失败:%s\n", err)
				continue
			}
			// Terminal size has not changed, don's do anything.
			if currTermHeight == c.win.h && currTermWidth == c.win.w {
				continue
			}
			err = c.session.WindowChange(currTermHeight, currTermWidth)
			if err != nil {
				log.Printf("Unable to send window-change reqest: %s\n", err)
				continue
			}
			c.win.w, c.win.h = currTermWidth, currTermHeight
		}
	}
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
	// 拿到当前终端文件描述符
	fd := int(os.Stdin.Fd())
	state, err := terminal.MakeRaw(fd)
	// 退出还原终端
	defer terminal.Restore(fd, state)
	if err != nil {
		log.Fatal("MakeRaw:", err)
	}
	// 终端大小;windows 下获取输出才能正确运行,目前linux和windows下获取输出调整窗口大小正常，暂时不做区分处理
	var ofd = int(os.Stdout.Fd())
	// 获取终端大小
	width, height, err := terminal.GetSize(ofd)
	if err != nil {
		log.Fatal("GetSize:", err)
	}
	c.win = &terminalWindow{
		h: height,
		w: width,
	}
	// 监听窗口变化
	go c.winChange(ofd)
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
	go func(w io.Writer) {
		buf := make([]byte, 128)
		for {
			n, err := os.Stdin.Read(buf)
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
	}(stdinPipe)

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
