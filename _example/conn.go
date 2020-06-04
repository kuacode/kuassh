package main

import "github.com/kuassh/cmd"

func main() {
	//n := &kuassh.Node{
	//	Host:     "127.0.0.1",
	//	Port:     "2222",
	//	User:     "root",
	//	PassWord: "admin",
	//	KeyFile:  "",
	//	Cmds:     nil,
	//}
	//newClient, _ := kuassh.NewClient(n)
	//newClient.Login()
	cmd.SSHExecute()
}
