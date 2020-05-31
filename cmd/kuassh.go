package cmd

import (
	"fmt"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"kuassh"
	"log"
	"os"
	"strings"
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
	node := selectNode(nil, nodes)
	c, err := kuassh.NewClient(node)
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
var templates = &promptui.SelectTemplates{
	Label:    "✨ {{ . | green}}",
	Active:   "➤ {{ .Name | cyan  }}{{if .Host}}{{if .User}}{{.User | faint}}{{`@` | faint}}{{end}}{{.Host | faint}}{{end}}",
	Inactive: "  {{.Name | faint}}{{if .Host}}{{if .User}}{{.User | faint}}{{`@` | faint}}{{end}}{{.Host | faint}}{{end}}",
}

// 光标位置
var cursor = 0

func clearSelectedBuf() {
	if cursor != 0 && cursor%2 == 0 {
		fmt.Print("\033[2A") // 上移2行
		fmt.Print("\x1b[2k") // 清除一行
		fmt.Print("\033[1B") // 下移一行
	} else {
		cursor += 1
	}
}

// 上级目录
const prev = "--parent--"

func selectNode(parent, nodes []*kuassh.Node) *kuassh.Node {
	// 终端选择 UI
	prompt := promptui.Select{
		Label:     "服务器列表",
		Items:     nodes,
		Templates: templates,
		Size:      20,
		//HideSelected: true, // 隐藏选择后顶部显示
		HistorySelectedCount: 2,
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
		return selectNode(nodes, node.Children)
	}
	if node.Name == prev {
		if parent == nil {
			return selectNode(nil, kuassh.GetConfig())
		}
		return selectNode(nil, parent)
	}
	return node
}
