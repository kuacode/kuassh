package sftp

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/sftp"
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

	var src string
	isAbs := filepath.IsAbs(args[1])
	if isAbs {
		src = args[1]
	} else {
		src = filepath.Join(sc.lWorkDir, filepath.Base(args[1]))
	}
	sinfo, err := os.Stat(src)
	if sinfo.IsDir() {
		filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
			if info.IsDir() {
				return sc.putCheckDir(strings.Replace(path, src, targetDir, 1), info.Mode())
			} else {
				return sc.upload(path, strings.Replace(path, src, targetDir, 1))
			}
		})
	} else {
		err = sc.putCheckDir(targetDir, 0644)
		if err != nil {
			fmt.Println("put -> check dir error:", err)
			return
		}
		sc.upload(src, sc.client.Join(targetDir, filepath.Base(src)))
	}
}

func (sc *sftpClient) putCheckDir(targetDir string, mode os.FileMode) error {
	_, err := sc.client.Stat(targetDir)
	if err != nil {
		if os.IsNotExist(err) {
			err = sc.client.MkdirAll(targetDir)
			if err != nil {
				fmt.Println("put -> making dir error")
				return err
			}
		}
	} else {
		return errors.New("文件夹已存在是否覆盖")
	}
	return err
}

func (sc *sftpClient) upload(src, target string) error {
	// local
	srcFile, err := os.Open(src)
	if err != nil {
		fmt.Println("put local file error:", err)
		return err
	}
	defer func(f *os.File) {
		_ = f.Close()
	}(srcFile)
	sinfo, err := srcFile.Stat()
	if err != nil {
		fmt.Println("stat local file error:", err)
		return err
	}
	// remote
	targetFile, err := sc.client.Create(target)
	if err != nil {
		fmt.Println("put target file error:", err)
		return err
	}
	defer func(f *sftp.File) {
		_ = f.Close()
	}(targetFile)
	sc.progressBar.NewBar(sinfo.Name(), sinfo.Size())
	sc.progressBar.pb.Start()
	io.Copy(io.MultiWriter(sc.progressBar, targetFile), srcFile)
	sc.progressBar.pb.Finish()
	//
	return sc.client.Chmod(target, sinfo.Mode())
}
