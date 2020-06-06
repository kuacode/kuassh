package kuassh

import (
	"fmt"
	"io/ioutil"
	"os/user"
	"path"
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
	b, err := LoadConfigBytes("kssh.yaml", ".kssh.yaml")
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

func fillValue(nodes []*Node) {
	for i, _ := range nodes {
		if len(nodes[i].Children) > 0 {
			fillValue(nodes[i].Children)
		}
		nodes[i].F = 1
		//
		nodes[i].NeedAuth = true
		if nodes[i].Port == "" {
			// 默认端口
			nodes[i].Port = "22"
		}
	}
}
