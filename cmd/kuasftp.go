package cmd

import (
	"fmt"
	"github.com/kuassh"
	kssh "github.com/kuassh/ssh"
	"github.com/mattn/go-tty"
	"github.com/pkg/sftp"
	"github.com/spf13/cobra"
	"log"
	"os"
)

var ksftpCmd = &cobra.Command{
	Use:   "ksftp",
	Short: "sftp",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		runSFTP()
	},
}

func SftpExecute() {
	if err := ksshCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func runSFTP() {
	err := kuassh.LoadConfig()
	if err != nil {
		log.Fatalln("加载配置错误:", err)
	}
	// 获取节点
	nodes := kuassh.GetConfig()
	//
	_tty, err := tty.Open()
	defer _tty.Close()
	if err != nil {
		log.Fatalln("tty创建错误:", err)
	}
	node := kuassh.SelectNode(nil, nodes)
	c, err := kssh.NewClient(node)
	if err != nil {
		log.Fatalln("获取客户端错误:", err)
	}
	sshClient := c.Login()
	// 创建sftp客户端
	sftp, err := sftp.NewClient(sshClient)
	defer sftp.Close()
	if err != nil {
		log.Fatal("创建sftp客户端错误", err)
	}
	// 监听窗口大小
	go func() {
		for ws := range _tty.SIGWINCH() {
			fmt.Println("Resized", ws.W, ws.H)
		}
	}()

	clean, err := _tty.Raw()
	if err != nil {
		log.Fatal(err)
	}
	defer clean()

	for {
		fmt.Print("SFTP =>")
		r, err := _tty.ReadRune()
		if err != nil {
			log.Fatal(err)
		}
		if r == 0 {
			continue
		}
		fmt.Println(r)
		if _tty.Buffered() {
			break
		}
	}
}

// 执行命令
func executeCommand(s string) {

}
