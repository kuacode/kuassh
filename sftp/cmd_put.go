package sftp

import (
	"fmt"
	"os"
	"path/filepath"
)

func (sc *sftpClient) put(args []string) {
	if len(args) < 2 {
		fmt.Println("put 缺少参数，get src | srcDir")
		return
	}
	if len(args) > 3 {
		fmt.Println("put 参数错误，put src | targetDir ... srcDir | targetDir")
		return
	}
	var targetDir string
	if len(args) == 3 {
		targetDir = args[2]
	} else {
		targetDir = sc.rWorkDir
	}
	_, err := sc.client.Stat(targetDir)
	if err != nil {
		if !os.IsNotExist(err) {
			fmt.Println("上传目录错误:", err)
			return
		}
		err = os.MkdirAll(targetDir, 0644)
		if err != nil {
			fmt.Println("get -> making dir error")
			return
		}
	}

	isAbs := filepath.IsAbs(args[0])
	if isAbs {

	} else {

	}
}
