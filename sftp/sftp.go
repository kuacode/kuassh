package sftp

import (
	"fmt"
	"github.com/c-bata/go-prompt/completer"
	"github.com/kuassh"
	"github.com/kuassh/pkg/go-prompt"
	kssh "github.com/kuassh/ssh"
	"github.com/pkg/sftp"
	"log"
	"os"
	"strings"
)

type sftpClient struct {
	client   *sftp.Client
	user     string
	rWorkDir string
	lWorkDir string
	out      chan int // 推出通道
}

func NewSftpClient() *sftpClient {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("获取用户主目录错误:", err)
	}
	return &sftpClient{
		lWorkDir: homeDir,
		rWorkDir: "~",
		out:      make(chan int),
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
	// 用户信息
	sc.user = sshClient.User()
	sc.rWorkDir, err = sc.client.Getwd()
	if err != nil {
		fmt.Printf("sftp获取远端workdir错误:%v\n", err)
		//
		sc.rWorkDir = "~"
	}
	sc.runTerminal()
}

func (sc *sftpClient) runTerminal() {
	// 推出
	go func() {
		<-sc.out
		err := sc.client.Close()
		if err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()

	p := prompt.New(
		sc.executor,
		sc.completer,
		prompt.OptionLivePrefix(sc.CreatePrompt),
		prompt.OptionInputTextColor(prompt.Green),
		prompt.OptionPrefixTextColor(prompt.Blue),
		prompt.OptionCompletionWordSeparator(completer.FilePathCompletionSeparator),
	)

	p.Run()
}

// 命令执行器
func (sc *sftpClient) executor(command string) {
	// 去除两边空格
	commands := splitCommand(command)
	// switch command
	switch commands[0] {
	case "bye", "exit", "quit":
		fmt.Println("Sftp Exit...")
		sc.out <- 1
	case "cd": // change remote directory
		sc.cd(commands)
	//case "chgrp":
	//	sc.chgrp(cmdline)
	//case "chmod":
	//	sc.chmod(cmdline)
	//case "chown":
	//	sc.chown(cmdline)
	case "": // none command...
	default:
		fmt.Println("Command Not Found...")
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
		{Text: "lmkdir", Description: "Create local directory"},
		// {Text: "ln", Description: "Link remote file (-s for symlink)"},
		{Text: "lpwd", Description: "Print local working directory"},
		{Text: "ls", Description: "Display remote directory listing"},
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

// 前缀
func (sc *sftpClient) CreatePrompt() (p string, result bool) {
	return "sftp >> ", true
}

func splitCommand(command string) []string {
	cmds := strings.Split(command, " ")
	var commands []string
	for _, cmd := range cmds {
		if cmd != " " {
			commands = append(commands, cmd)
		}
	}
	return commands
}
