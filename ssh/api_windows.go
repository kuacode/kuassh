// +build windows

package ssh

import (
	"log"
	"time"

	"github.com/containerd/console"
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

// func (c *client) winChange(t *tty.TTY) {
// 	// 监听窗口
// 	for ws := range t.SIGWINCH() {
// 		if c.win.w != ws.W || c.win.h != ws.H {
// 			err := c.session.WindowChange(ws.H, ws.W)
// 			if err != nil {
// 				log.Printf("调整窗口大小错误:%v\n", err)
// 			} else {
// 				c.win.w, c.win.h = ws.W, ws.H
// 			}
// 		}
// 	}
// }

// 监听窗口大小变化
func (c *client) winChange(current console.Console) {
	t := time.Tick(time.Second)
	for range t {
		ws, err := current.Size()
		if err != nil {
			log.Printf("获取当前窗口大小失败:%s\n", err)
			continue
		}
		currTermWidth, currTermHeight := int(ws.Width), int(ws.Height)
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
