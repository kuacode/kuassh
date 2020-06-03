package cmd

import (
	"fmt"
	"github.com/kuassh"
	"github.com/kuassh/ssh"
	"github.com/mattn/go-tty"
	"log"
	"os"
	"strings"

	"github.com/kuassh/pkg/promptui"
	"github.com/spf13/cobra"
)

var ksshCmd = &cobra.Command{
	Use:   "kssh",
	Short: "终端管理器",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		run()
	},
}

func Execute() {
	if err := ksshCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run() {
	err := kuassh.LoadConfig([]string{"kssh.yaml", ".kssh.yaml"})
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
	node := selectNode(nil, nodes, _tty)
	c, err := ssh.NewClient(node)
	if err != nil {
		log.Fatalln("获取客户端错误:", err)
	}
	c.Login()
}

//var templates = &promptui.SelectTemplates{
//	Label:    " ✨ {{ . | green }}",
//	Active:   "\U0001F336 {{ .Name | green }} ({{ .User | faint }}@{{ .Host | faint }})",
//	Inactive: "  {{ .Name | cyan }} ({{ .User | faint }}@{{ .Host | faint }})",
//	Selected: "\U0001F336 {{ .Host | green }}",
//	Details: `
//--------- 详细 ----------
//{{ "Name:" | faint }}	{{ .Name }}
//{{ "Host:" | faint }}	{{ .Host }}
//{{ "User:" | faint }}	{{ .User }}
//{{ "Port:" | faint }}	{{ .Port }}
//`,
//}
var (
	templates = &promptui.SelectTemplates{
		Label:    "✨ {{ . | green}}",
		Active:   "➤ {{ .Name | cyan  }}{{if .Host}}{{if .User}}{{.User | faint}}{{`@` | faint}}{{end}}{{.Host | faint}}{{end}}",
		Inactive: "  {{.Name | faint}}{{if .Host}}{{if .User}}{{.User | faint}}{{`@` | faint}}{{end}}{{.Host | faint}}{{end}}",
		Selected: "\U0001F336 {{ .Name | green }}",
	}
)

// 上级目录
const prev = "--parent--"

func selectNode(parent, nodes []*kuassh.Node, t *tty.TTY) *kuassh.Node {
	// 终端选择 UI
	prompt := promptui.Select{
		Label:     "服务器列表",
		Items:     nodes,
		Templates: templates,
		Size:      20,
		//HideSelected: true, // 隐藏选择后顶部显示
		HistorySelectedCount: 2,
		Stdin:                t.Input(),
		Stdout:               t.Output(),
		Searcher: func(input string, index int) bool {
			n := nodes[index]
			content := fmt.Sprintf("%s %s %s", n.Name, n.User, n.Host)
			if strings.Contains(content, input) {
				return true
			}
			// 多个匹配
			if strings.Contains(input, " ") {
				for _, sp := range strings.Split(input, " ") {
					sp = strings.TrimSpace(sp)
					if sp != "" {
						if !strings.Contains(content, sp) {
							return false
						}
					}
				}
				return true
			}
			return false
		},
	}
	index, _, err := prompt.Run()
	if err != nil {
		// 退出不输出
		if err == promptui.ErrInterrupt || err == promptui.ErrEOF || err == promptui.ErrAbort {
			os.Exit(0)
		}
		log.Fatal("终端选择节点错误", err)
	}
	node := nodes[index]
	// 子节点
	if len(node.Children) > 0 {
		first := node.Children[0]
		if first.Name != prev {
			// 创建一个返回上一级节点
			prevNode := &kuassh.Node{Name: prev}
			node.Children = append([]*kuassh.Node{prevNode}, node.Children...)
		}
		return selectNode(nodes, node.Children, t)
	}
	if node.Name == prev {
		if parent == nil {
			return selectNode(nil, kuassh.GetConfig(), t)
		}
		return selectNode(nil, parent, t)
	}
	return node
}
