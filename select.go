package kuassh

import (
	"fmt"
	"github.com/kuassh/pkg/promptui"
	"log"
	"os"
	"strings"
)

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
		Active:   "➤ 🟢{{if .Children}} 📁{{else if eq .Name `--parent--`}} ✈{{else}} 🚀{{end}}{{ .Name | faint  }} {{if .Host}}{{if .User}}{{.User | faint}}{{`@` | faint}}{{end}}{{.Host | faint}}{{end}}",
		Inactive: "  🟡{{if .Children}} 📁{{else if eq .Name `--parent--`}} ✈{{else}} 🚀{{end}}{{ .Name | faint}} {{if .Host}}{{if .User}}{{.User | faint}}{{`@` | faint}}{{end}}{{.Host | faint}}{{end}}",
		Selected: "\U0001F336{{.Name | green }}",
	}
)

// 上级目录
const prev = "--parent--"

func SelectNode(parent, nodes []*Node) *Node {
	// 终端选择 UI
	prompt := promptui.Select{
		Label:     "服务器列表",
		Items:     nodes,
		Templates: templates,
		Size:      20,
		//HideSelected: true, // 隐藏选择后顶部显示
		HistorySelectedCount: -1,
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
			prevNode := &Node{Name: prev, F: 2}
			node.Children = append([]*Node{prevNode}, node.Children...)
		}
		return SelectNode(nodes, node.Children)
	}
	if node.Name == prev {
		if parent == nil {
			return SelectNode(nil, GetConfig())
		}
		return SelectNode(nil, parent)
	}
	return node
}
