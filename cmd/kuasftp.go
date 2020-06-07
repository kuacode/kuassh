package cmd

import (
	"fmt"
	"github.com/kuassh"
	ksftp "github.com/kuassh/sftp"
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
	if err := ksftpCmd.Execute(); err != nil {
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
	node := kuassh.SelectNode(nil, nodes)
	sc := ksftp.NewSftpClient()
	sc.Login(node)
}
