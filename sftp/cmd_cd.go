package sftp

import (
	"fmt"
)

func (sc *sftpClient) cd(args []string) {
	// cd command only
	if len(args) == 1 {
		sc.rWorkDir = "~"
		return
	}
	rdir := args[1]
	if rdir[0] != '/' { // 全路径
		rdir = sc.client.Join(sc.rWorkDir, rdir)
	}
	//rp, err := sc.client.ReadLink(rdir)
	//if err != nil {
	//	fmt.Printf("Error: %s\n", err)
	//	return
	//}
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
	sc.client.Walk(rdir)
}
