package ssh

import (
	"github.com/kuassh"
	"testing"
)

func TestNewClient(t *testing.T) {
	n := &kuassh.Node{
		Host:     "127.0.0.1",
		Port:     "22",
		User:     "root",
		PassWord: "123456",
		KeyFile:  "",
		Cmds:     nil,
	}
	newClient, _ := NewClient(n)
	newClient.Login()
}
