package sftp

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
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
	if rfInfo.IsDir() {
		// local file
		downloadDir := filepath.Join(downloadDir, filepath.Base(rdir))
		err := sc.checkDir(downloadDir, rfInfo.Mode())
		if err != nil {
			fmt.Println("get check dir error:", err)
			return
		}
		//
		w := sc.client.Walk(rdir)
		for w.Step() {
			// skip
			if w.Path() == rdir {
				continue
			}
			if w.Stat().IsDir() {
				err = os.Mkdir(strings.Replace(w.Path(), rdir, downloadDir, -1), w.Stat().Mode())
				if err != nil {
					fmt.Println("get making dir error:", err)
					continue
				}
			} else {
				target := strings.Replace(w.Path(), rdir, downloadDir, 1)
				sc.download(w.Path(), target, w.Stat().Mode())
			}
		}
	} else { // remote is file
		err := sc.checkDir(downloadDir, 0644) // drw--w--w-
		if err != nil {
			fmt.Println("get check dir error:", err)
			return
		}
		target := filepath.Join(downloadDir, rfInfo.Name())
		sc.download(rdir, target, rfInfo.Mode())
	}
}

func (sc *sftpClient) checkDir(targetDir string, mode os.FileMode) error {
	_, err := os.Stat(targetDir)
	// if local dir not exist, we will create a local dir
	// with the same name as the remote dir
	if err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(targetDir, mode)
			if err != nil {
				fmt.Println("get -> making dir error")
				return err
			}
		}
	} else {
		return errors.New("文件夹已存在是否覆盖")
	}
	return nil
}

func (sc *sftpClient) download(src, target string, fm os.FileMode) {
	// if remote path is a file, copy it
	targetFile, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, fm)
	if err != nil {
		fmt.Println("open local file error:", err)
		return
	}
	defer targetFile.Close()
	//
	srcFile, err := sc.client.Open(src)
	if err != nil {
		fmt.Println("open remote file error:", err)
		return
	}
	defer srcFile.Close()
	//
	srcFileInfo, _ := srcFile.Stat()
	sc.NewProcessBar(srcFileInfo.Name(), srcFileInfo.Size())
	sc.pb.Start()
	//
	buf := make([]byte, 32*1024)
	for {
		n, err := srcFile.Read(buf)
		if n > 0 {
			sc.pb.Add(n)
			targetFile.Write(buf[:n])
		}
		if err != nil {
			if err != io.EOF {
				fmt.Println("downloading remote file error:", err)
			}
			break
		}
	}
	sc.pb.Finish()
}
