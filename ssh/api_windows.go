// +build windows

package ssh

import (
	"github.com/mattn/go-tty"
	"io"
	"log"
)

//
//// 监听窗口大小变化
//func (c *client) winChange(fd int) {
//	t := time.Tick(time.Second)
//	for range t {
//		currTermWidth, currTermHeight, err := terminal.GetSize(fd)
//		if err != nil {
//			log.Printf("获取当前窗口大小失败:%s\n", err)
//			continue
//		}
//		// 窗口大小发生变化
//		if currTermHeight == c.win.h && currTermWidth == c.win.w {
//			continue
//		}
//		err = c.session.WindowChange(currTermHeight, currTermWidth)
//		if err != nil {
//			log.Printf("Unable to send window-change reqest: %s\n", err)
//			continue
//		}
//		c.win.w, c.win.h = currTermWidth, currTermHeight
//	}
//}

func (c *client) winChange(t *tty.TTY) {
	// 监听窗口
	for ws := range t.SIGWINCH() {
		if c.win.w != ws.W || c.win.h != ws.H {
			err := c.session.WindowChange(ws.H, ws.W)
			if err != nil {
				log.Printf("调整窗口大小错误:%v\n", err)
			} else {
				c.win.w, c.win.h = ws.W, ws.H
			}
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
	// tty
	t, err := tty.Open()
	if err != nil {
		log.Fatal("tty:", err)
	}
	// 还原终端？
	clean, err := t.Raw()
	if err != nil {
		log.Fatal("tty:", err)
	}
	defer clean()
	// win size
	width, height, err := t.Size()
	c.win = &terminalWindow{h: height, w: width}
	// 监听窗口变化
	go c.winChange(t)
	// 请求Pty
	c.requestPty()
	// 直接对接了 stderr、stdout 和 stdin 会造成 tmux等出问题 ，实际上我们应当启动一个异步的管道式复制行为
	stdoutPipe, err := c.session.StdoutPipe()
	if err != nil {
		log.Fatal("StdoutPipe", err)
	}
	go func(r io.Reader) {
		_, _ = io.Copy(t.Output(), r)
	}(stdoutPipe)
	//
	stderrPipe, err := c.session.StderrPipe()
	if err != nil {
		log.Fatal("StderrPipe", err)
	}
	go func(r io.Reader) {
		_, _ = io.Copy(t.Output(), r)
	}(stderrPipe)
	//
	stdinPipe, err := c.session.StdinPipe()
	if err != nil {
		log.Fatal("StdinPipe", err)
	}
	// 系统终端输入拷贝到远程终端执行
	go func(w io.Writer) {
		for {
			r, _ := t.ReadRune()
			if r == 0 {
				continue
			}
			stdinPipe.Write([]byte(string(r)))
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
