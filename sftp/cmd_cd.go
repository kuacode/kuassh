package sftp

import (
	"fmt"
	"os"
	"path/filepath"
)

func (sc *sftpClient) cd(args []string) {
	// cd command only
	if len(args) == 1 {
		sc.rWorkDir = sc.rUserHome
		return
	}
	rdir := args[1]
	if rdir[0] != '/' { // 全路径
		rdir = sc.client.Join(sc.rWorkDir, rdir)
	}
	// get stat
	stat, err := sc.client.Lstat(rdir)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}
	if !stat.IsDir() {
		fmt.Printf("Error: %s\n", "is not directory")
		return
	}
	sc.rWorkDir = args[1]
	// sc.client.Walk(rdir)
}

func (sc *sftpClient) lcd(args []string) {
	// cd command only
	if len(args) == 1 {
		sc.lWorkDir = sc.lUserHome
		return
	}
	ldir := args[1]
	if !filepath.IsAbs(ldir) { // 全路径
		ldir = filepath.Join(sc.lWorkDir, ldir)
	}
	// get stat
	stat, err := os.Stat(ldir)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}
	if !stat.IsDir() {
		fmt.Printf("Error: %s\n", "is not directory")
		return
	}
	sc.lWorkDir = args[1]
	// sc.client.Walk(rdir)
}
