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
