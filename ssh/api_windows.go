// +build windows

package ssh

import (
	"golang.org/x/crypto/ssh/terminal"
	"log"
	"time"
)

// 监听窗口大小变化
func (c *client) winChange(fd int) {
	t := time.Tick(time.Second)
	for range t {
		currTermWidth, currTermHeight, err := terminal.GetSize(fd)
		if err != nil {
			log.Printf("获取当前窗口大小失败:%s\n", err)
			continue
		}
		// 窗口大小发生变化
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
