package sftp

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/fatih/color"
)

func (sc *sftpClient) ls(cmds []string) {
	sc.printFiles(true, cmds)
}

func (sc *sftpClient) lls(cmds []string) {
	sc.printFiles(false, cmds)
}

func (sc *sftpClient) printFiles(r bool, cmds []string) {
	var (
		fs   []os.FileInfo
		err  error
		a, l bool
	)
	if r {
		fs, err = sc.client.ReadDir(sc.rWorkDir)
	} else {
		fs, err = ioutil.ReadDir(sc.lWorkDir)
	}
	if err != nil {
		fmt.Printf("获取文件列表错误:%s-%v", sc.rWorkDir, err)
		return
	}

	if len(cmds) == 1 {
		if cmds[0] == "ls" || cmds[0] == "lls" {
			a, l = false, false
		} else {
			a, l = false, true
		}
	} else if len(cmds) == 2 {
		if cmds[1] == "-a" {
			a, l = true, false
		} else if cmds[1] == "-l" {
			a, l = false, true
		} else if cmds[1] == "-la" || cmds[1] == "-al" {
			a, l = true, true
		} else {
			fmt.Printf("命令错误，请重新输入\n")
		}
	}
	//
	var s = "%s\t"
	if l {
		s = "%s\n"
	}
	for _, f := range fs {
		if !a && strings.HasPrefix(f.Name(), ".") {
			continue
		}
		if l {
			fmt.Printf("%s %d ", f.Mode(), f.Size())
		}
		switch {
		case f.IsDir():
			fmt.Printf(s, color.BlueString("%s", f.Name()))
		case strings.Count(f.Mode().String(), "x") == 3:
			fmt.Printf(s, color.GreenString("%s", f.Name()))
		default:
			fmt.Printf(s, color.WhiteString("%s", f.Name()))
		}
	}
	fmt.Println()
}
