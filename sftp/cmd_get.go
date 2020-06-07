package sftp

import (
	"fmt"
	"github.com/cheggaaa/pb/v3"
	"io"
	"os"
	"path"
	"strings"
)

// 下载
func (sc *sftpClient) get(args []string) {
	if len(args) < 2 {
		fmt.Println("get 缺少参数，get src | srcDir")
		return
	}
	if len(args) > 3 {
		fmt.Println("get 缺少参数，get src | srcDir ... target | targetDir")
		return
	}
	var downloadDir string
	if len(args) == 3 {
		downloadDir = args[2]
	} else {
		downloadDir = sc.lWorkDir
	}

	rdir := args[1]
	if rdir[0] != '/' { // 全路径
		rdir = sc.client.Join(sc.rWorkDir, rdir)
	}
	rf, err := sc.client.Open(rdir)
	defer rf.Close()
	if err != nil {
		fmt.Println("get error:", err)
		return
	}
	rfInfo, err := rf.Stat()
	if err != nil {
		fmt.Println("get error:", err)
		return
	}
	// local file
	lf, lerr := os.OpenFile(downloadDir, os.O_RDWR|os.O_CREATE, 0644)
	defer lf.Close()
	lfInfo, err := lf.Stat()

	if rfInfo.IsDir() {
		// if local dir not exist, we will create a local dir
		// with the same name as the remote dir
		if lerr != nil {
			if os.IsNotExist(lerr) {
				lerr = os.MkdirAll(path.Join(downloadDir), rfInfo.Mode())
				if lerr != nil {
					fmt.Println("get making dir error:", lerr)
				}
			}
		} else {
			// if local dir exist, we will merge local dir path and remote dir relative path
			if lfInfo.IsDir() {
				// create local dir
				downloadDir = path.Join(downloadDir, path.Base(rdir))
				gerr := os.MkdirAll(downloadDir, rfInfo.Mode())
				if gerr != nil {
					fmt.Println("get making dir error:", gerr)
				}
			}
		}
		//
		w := sc.client.Walk(rdir)
		for w.Step() {
			// skip
			if w.Path() == rdir {
				continue
			}
			if w.Stat().IsDir() {
				err = os.Mkdir(strings.Replace(w.Path(), rdir, downloadDir, -1), rfInfo.Mode())
				if err != nil {
					fmt.Println("get making dir error:", err)
				}
			} else {
				// if remote path is a file, copy it
				localFile, err := os.OpenFile(strings.Replace(w.Path(), rdir, downloadDir, 1), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, w.Stat().Mode())
				if err != nil {
					fmt.Println("open local file error:", err)
				}
				remoteTmpFile, err := sc.client.Open(w.Path())
				if err != nil {
					fmt.Println("open remote file error:", err)
				}
				//
				rfTempInfo, _ := remoteTmpFile.Stat()
				sc.pb = pb.New64(rfTempInfo.Size())
				sc.pb.Start()
				//
				buf := make([]byte, 32*1024)
				for {
					n, err := remoteTmpFile.Read(buf)
					if n > 0 {
						sc.pb.Add(n)
					}
					if err != nil {
						if err != io.EOF {
							fmt.Println("downloading remote file error:", err)
						}
						break
					}
				}
				sc.pb.Finish()
				// 关闭
				_ = remoteTmpFile.Close()
				_ = localFile.Close()
			}
		}
	}
}
