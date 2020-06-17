package sftp

import (
	"fmt"
	"path"
	"strings"
)

func (sc *sftpClient) mkdir(args []string) {
	lenArgs := len(args)
	if lenArgs == 1 || lenArgs > 2 {
		fmt.Printf("mkdir dirname")
		return
	}
	//
	var target string
	if strings.HasPrefix(args[1], "/") {
		target = args[1]
	} else {
		target = path.Join(sc.rWorkDir, args[1])
	}
	sc.client.Mkdir(target)
}
