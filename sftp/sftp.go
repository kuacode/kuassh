package sftp

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/cheggaaa/pb/v3"
	"github.com/chzyer/readline"
	"github.com/kuassh"
	"github.com/kuassh/pkg/go-prompt"
	kssh "github.com/kuassh/ssh"
	"github.com/pkg/sftp"
)

///  tmpl := `{{ red "With funcs:" }} {{ bar . "<" "-" (cycle . "↖" "↗" "↘" "↙" ) "." ">"}} {{speed . | rndcolor }} {{percent .}} {{string . "my_green_string" | green}} {{string . "my_blue_string" | blue}}`
//// start bar based on our template
//   bar := pb.ProgressBarTemplate(tmpl).Start64(limit)
//// set values for string elements
//   bar.Set("my_green_string", "green").
//	 Set("my_blue_string", "blue")

type sftpClient struct {
	client *sftp.Client
	// 远程
	rUser     string
	rUserHome string
	rWorkDir  string
	// 本地
	lUserHome string
	lWorkDir  string
	pb        *pb.ProgressBar
	Instance  *readline.Instance
}

func NewSftpClient() *sftpClient {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("获取用户主目录错误:", err)
	}
	return &sftpClient{
		lUserHome: homeDir,
		lWorkDir:  homeDir,
	}
}

func (sc *sftpClient) Login(node *kuassh.Node) {
	c, err := kssh.NewClient(node)
	if err != nil {
		log.Fatal("获取客户端错误:", err)
	}
	sshClient := c.Login()
	// 创建sftp客户端
	sc.client, err = sftp.NewClient(sshClient)
	if err != nil {
		log.Fatal("创建sftp客户端错误", err)
	}
	//用户信息
	sc.rUser = sshClient.User()
	sc.rUserHome, err = sc.client.Getwd()
	if err != nil {
		fmt.Printf("sftp获取远端workdir错误:%v\n", err)
		//
		sc.rWorkDir = "~"
	}
	sc.rWorkDir = sc.rUserHome
	sc.runTerminal()
}

func (sc *sftpClient) runTerminal() {
	p := prompt.New(
		sc.executor,
		sc.completer,
		prompt.OptionPrefix(">>> "),
	)
	p.Run()
}

func splitCommand(command string) []string {
	cmds := strings.Split(strings.Trim(command, " "), " ")
	var commands []string
	for _, cmd := range cmds {
		if cmd != " " {
			commands = append(commands, cmd)
		}
	}
	return commands
}

func (sc *sftpClient) executor(line string) {
	cmds := splitCommand(line)
	switch {
	case line == "":
	case cmds[0] == "login":
		// todo
	case cmds[0] == "bye":
		os.Exit(0)
	case cmds[0] == "pwd":
		println(sc.rWorkDir)
	case cmds[0] == "lpwd":
		println(sc.lWorkDir)
	case cmds[0] == "cd": // change remote directory
		sc.cd(cmds)
	case cmds[0] == "lcd": // change local directory
		sc.lcd(cmds)
	case cmds[0] == "ls" || cmds[0] == "ll":
		sc.ls(cmds)
	case cmds[0] == "lls" || cmds[0] == "lll":
		sc.lls(cmds)
	case cmds[0] == "get":
		sc.get(cmds)
	case cmds[0] == "sleep":
		fmt.Println("sleep 4 second")
		time.Sleep(4 * time.Second)
	default:
		fmt.Println("命令错误:", strconv.Quote(line))
	}
}

// 提示函数
func (sc *sftpClient) completer(d prompt.Document) []prompt.Suggest {
	s := []prompt.Suggest{
		{Text: "bye", Description: "Quit lsftp"},
		// {Text: "cat", Description: "Open file"},
		{Text: "cd", Description: "Change remote directory to 'path'"},
		{Text: "chgrp", Description: "Change group of file 'path' to 'grp'"},
		{Text: "chown", Description: "Change owner of file 'path' to 'own'"},
		// {Text: "copy", Description: "Copy to file from 'remote' or 'local' to 'remote' or 'local'"},
		{Text: "df", Description: "Display statistics for current directory or filesystem containing 'path'"},
		{Text: "exit", Description: "Quit lsftp"},
		{Text: "get", Description: "Download file"},
		// {Text: "reget", Description: "Resume download file"},
		// {Text: "reput", Description: "Resume upload file"},
		{Text: "help", Description: "Display this help text"},
		{Text: "lcd", Description: "Change local directory to 'path'"},
		{Text: "lls", Description: "Display local directory listing"},
		{Text: "lll", Description: "Display local directory listing"},
		{Text: "lmkdir", Description: "Create local directory"},
		// {Text: "ln", Description: "Link remote file (-s for symlink)"},
		{Text: "lpwd", Description: "Print local working directory"},
		{Text: "ls", Description: "Display remote directory listing"},
		{Text: "ll", Description: "Display remote directory listing"},
		// {Text: "lumask", Description: "Set local umask to 'umask'"},
		{Text: "mkdir", Description: "Create remote directory"},
		// {Text: "progress", Description: "Toggle display of progress meter"},
		{Text: "put", Description: "Upload file"},
		{Text: "pwd", Description: "Display remote working directory"},
		{Text: "quit", Description: "Quit sftp"},
		{Text: "rename", Description: "Rename remote file"},
		{Text: "rm", Description: "Delete remote file"},
		{Text: "rmdir", Description: "Remove remote directory"},
		{Text: "symlink", Description: "Create symbolic link"},
		// {Text: "tree", Description: "Tree view remote directory"},
		// {Text: "!command", Description: "Execute 'command' in local shell"},
		{Text: "!", Description: "Escape to local shell"},
		{Text: "?", Description: "Display this help text"},
	}
	return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)
}
