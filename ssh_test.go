package kuassh

import "testing"

func TestNewClient(t *testing.T) {
	n := &Node{
		Host:     "149.28.25.177",
		Port:     "22",
		User:     "root",
		PassWord: "%9aA-jR1[973FBn$",
		KeyFile:  "",
		Cmds:     nil,
	}
	newClient, _ := NewClient(n)
	newClient.Login()
}
