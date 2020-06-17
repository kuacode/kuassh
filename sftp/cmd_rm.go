package sftp

import (
	"fmt"
	"path"
	"strings"
)

func (sc *sftpClient) rm(args []string) {
	lenArgs := len(args)
	if lenArgs == 1 || lenArgs > 2 {
		fmt.Printf("rm file")
		return
	}
	//
	var target string
	if strings.HasPrefix(args[1], "/") {
		target = args[1]
	} else {
		target = path.Join(sc.rWorkDir, args[1])
	}
	err := sc.client.Remove(target)
	if err != nil {
		fmt.Println("rm file error:", err)
	}
}

func (sc *sftpClient) rmdir(args []string) {
	lenArgs := len(args)
	if lenArgs == 1 || lenArgs > 2 {
		fmt.Printf("rmdir dir")
		return
	}
	//
	var target string
	if strings.HasPrefix(args[1], "/") {
		target = args[1]
	} else {
		target = path.Join(sc.rWorkDir, args[1])
	}
	err := sc.client.Remove(target)
	if err != nil {
		fmt.Println("rmdir dir error:", err)
	}
}
