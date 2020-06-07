package sftp

import (
	"fmt"
	"github.com/cheggaaa/pb/v3"
	"github.com/chzyer/readline"
	"github.com/kuassh"
	"github.com/kuassh/pkg/go-prompt"
	kssh "github.com/kuassh/ssh"
	"github.com/pkg/sftp"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
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
	conf := &readline.Config{
		Prompt:          "\033[31m»\033[0m ",
		HistoryFile:     "/tmp/readline.tmp",
		AutoComplete:    completer,
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",

		HistorySearchFold:   true,
		FuncFilterInputRune: filterInput,
	}

	l, err := readline.NewEx(conf)
	if err != nil {
		panic(err)
	}
	defer l.Close()
	log.SetOutput(l.Stderr())
	for {
		line, err := l.Readline()
		if err == readline.ErrInterrupt {
			if len(line) == 0 {
				break
			} else {
				continue
			}
		} else if err == io.EOF {
			break
		}
		cmds := splitCommand(line)
		switch {
		case line == "":
		case cmds[0] == "login":
			pswd, err := l.ReadPassword("please enter your password: ")
			if err != nil {
				break
			}
			println("you enter:", strconv.Quote(string(pswd)))
		case cmds[0] == "bye":
			goto exit
		case cmds[0] == "pwd":
			println(sc.rWorkDir)
		case cmds[0] == "lpwd":
			println(sc.lWorkDir)
		case cmds[0] == "cd": // change remote directory
			sc.cd(cmds)
		case cmds[0] == "lcd": // change local directory
			sc.lcd(cmds)
		case cmds[0] == "ls":
			sc.ls(cmds)
		case cmds[0] == "lls":
			sc.lls(cmds)
		case cmds[0] == "get":
			sc.get(cmds)
		case cmds[0] == "help":
			usage(l.Stderr())
		case cmds[0] == "sleep":
			log.Println("sleep 4 second")
			time.Sleep(4 * time.Second)
		default:
			log.Println("命令错误:", strconv.Quote(line))
		}
	}
exit:
}

// 监控输出
func filterInput(r rune) (rune, bool) {
	switch r {
	// block CtrlZ feature
	case readline.CharCtrlZ:
		return r, false
	}
	return r, true
}

// 帮助信息
func usage(w io.Writer) {
	io.WriteString(w, "commands:\n")
	io.WriteString(w, completer.Tree("    "))
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
