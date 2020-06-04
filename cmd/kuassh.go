package cmd

import (
	"fmt"
	"github.com/kuassh"
	kssh "github.com/kuassh/ssh"
	"github.com/mattn/go-tty"
	"github.com/spf13/cobra"
	"log"
	"os"
)

var ksshCmd = &cobra.Command{
	Use:   "kssh",
	Short: "终端管理器",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		runSSH()
	},
}

func SSHExecute() {
	if err := ksshCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func runSSH() {
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
	node := kuassh.SelectNode(nil, nodes, _tty)
	c, err := kssh.NewClient(node)
	if err != nil {
		log.Fatalln("获取客户端错误:", err)
	}
	c.Login()
	// 开始会话
	c.StartSession()
}
