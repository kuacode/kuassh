package kuassh

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

type Node struct {
	Name     string      `yaml:"name"`
	Host     string      `yaml:"host"`
	Port     string      `yaml:"port"`
	User     string      `yaml:"user"`
	PassWord string      `yaml:"password"`
	KeyFile  string      `yaml:"keypath"`
	NeedAuth bool        `yaml:"needauth"`
	Jump     []*Node     `yaml:"jump"`
	Cmds     []*ShellCmd `yaml:"cmds"`
	Children []*Node     `yaml:"children"`
	// 上级，下级
	F int // 1 back 2 forward
}
type ShellCmd struct {
	Cmd   string        `yaml:"cmd"`
	Delay time.Duration `yaml:"delay"`
}

var (
	configs []*Node
)

func (n *Node) String() string {
	return fmt.Sprintf("%s@%s:%s", n.User, n.Host, n.Port)
}

func GetConfig() []*Node {
	return configs
}

func LoadConfig() error {
	execPath := os.Args[0]
	var (
		b   []byte
		err error
	)
	if execPath != "" {
		ss := strings.Split(execPath, string(filepath.Separator))
		LoadConfigBytes(strings.Join(append(ss[:len(ss)-1], "kssh.yaml"), string(filepath.Separator)), strings.Join(append(ss[:len(ss)-1], ".kssh.yaml"), string(filepath.Separator)))
	} else {
		b, err = LoadConfigBytes("kssh.yaml", ".kssh.yaml")
	}
	if err != nil {
		return err
	}
	nodes := []*Node{}
	err = yaml.Unmarshal(b, &nodes)
	if err != nil {
		return err
	}
	fillValue(nodes)
	configs = nodes
	return nil
}

func LoadConfigBytes(names ...string) ([]byte, error) {
	u, err := user.Current()
	if err != nil {
		return nil, err
	}
	// 用户主目录
	for i := range names {
		kssh, err := ioutil.ReadFile(path.Join(u.HomeDir, names[i]))
		if err == nil {
			return kssh, nil
		}
	}
	// 相对路径
	for i := range names {
		kssh, err := ioutil.ReadFile(names[i])
		if err == nil {
			return kssh, nil
		}
	}
	return nil, err
}

// 上级目录
const prev = "上一级"

func fillValue(nodes []*Node) {
	for i, _ := range nodes {
		if len(nodes[i].Children) > 0 {
			// 创建一个返回上一级节点
			prevNode := &Node{Name: prev, F: 2}
			nodes[i].Children = append([]*Node{prevNode}, nodes[i].Children...)
			fillValue(nodes[i].Children)
		}
		if nodes[i].F != 2 {
			nodes[i].F = 1
		}
		//
		nodes[i].NeedAuth = true
		if nodes[i].Port == "" && nodes[i].Host != "" {
			// 默认端口
			nodes[i].Port = "22"
		}
	}
}
