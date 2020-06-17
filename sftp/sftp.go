package sftp

import (
	"fmt"
	"github.com/c-bata/go-prompt/completer"
	"github.com/cheggaaa/pb/v3"
	"github.com/kuassh"
	"github.com/kuassh/pkg/go-prompt"
	kssh "github.com/kuassh/ssh"
	"github.com/pkg/sftp"
	"log"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

///  tmpl := `{{ red "With funcs:" }} {{ bar . "<" "-" (cycle . "↖" "↗" "↘" "↙" ) "." ">"}} {{speed . | rndcolor }} {{percent .}} {{string . "my_green_string" | green}} {{string . "my_blue_string" | blue}}`
//// start bar based on our template
//   bar := pb.ProgressBarTemplate(tmpl).Start64(limit)
//// set values for string elements
//   bar.Set("my_green_string", "green").
//	 Set("my_blue_string", "blue")
var tmpl = `{{ counters . "%s/%s" "%s/?"}} {{ bar . "[" "=" (cycle . ">" ) "." "]"}} {{speed .}} {{percent .}} {{rtime . "%s" "%s" "???"}}`

type sftpClient struct {
	client *sftp.Client
	// 远程
	rUser     string
	rUserHome string
	rWorkDir  string
	// PathComplete
	rComplete []prompt.Suggest
	/// 本地
	lUserHome   string
	lWorkDir    string
	progressBar *bar
	// PathComplete
	lComplete []prompt.Suggest
}

func NewSftpClient() *sftpClient {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("获取用户主目录错误:", err)
	}
	return &sftpClient{
		lUserHome: homeDir,
		lWorkDir:  homeDir,
		// progressBar
		progressBar: new(bar),
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
		prompt.OptionCompletionWordSeparator(completer.FilePathCompletionSeparator),
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
	switch cmds[0] {
	case "login":
		// todo
	case "bye", "exit", "quit":
		os.Exit(0)
	case "pwd":
		println(sc.rWorkDir)
	case "lpwd":
		println(sc.lWorkDir)
	case "cd": // change remote directory
		sc.cd(cmds)
	case "lcd": // change local directory
		sc.lcd(cmds)
	case "ls", "ll":
		sc.ls(cmds)
	case "lls", "lll":
		sc.lls(cmds)
	case "get":
		sc.get(cmds)
	case "put":
		sc.put(cmds)
	case "rm":
		sc.rm(cmds)
	case "rmdir":
		sc.rmdir(cmds)
	case "mkdir":
		sc.mkdir(cmds)
	default:
		fmt.Println("命令错误:", strconv.Quote(line))
	}
}

// 提示函数
func (sc *sftpClient) completer(d prompt.Document) []prompt.Suggest {
	// result
	var suggest []prompt.Suggest
	// Get cursor left
	left := d.CurrentLineBeforeCursor()
	// Get cursor char(string)
	char := ""
	if len(left) > 0 {
		char = string(left[len(left)-1])
	}
	cmds := strings.Split(left, " ")
	if len(cmds) == 1 {
		suggest = []prompt.Suggest{
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
	} else {
		switch cmds[0] {
		case "cd":
			return sc.PathComplete(true, 1, d)
		case "chgrp":
			// TODO(blacknon): そのうち追加 ver0.6.1
		case "chown":
			// TODO(blacknon): そのうち追加 ver0.6.1
		case "df":
			suggest = []prompt.Suggest{
				{Text: "-h", Description: "print sizes in powers of 1024 (e.g., 1023M)"},
				{Text: "-i", Description: "list inode information instead of block usage"},
			}
			return prompt.FilterHasPrefix(suggest, d.GetWordBeforeCursor(), false)
		case "get":
			switch {
			case strings.Count(d.CurrentLineBeforeCursor(), " ") == 1: // remote
				return sc.PathComplete(true, 1, d)
			case strings.Count(d.CurrentLineBeforeCursor(), " ") == 2: // local
				return sc.PathComplete(false, 2, d)
			}

		case "lcd":
			return sc.PathComplete(false, 1, d)
		case "lls":
			// switch options or path
			switch {
			case "-" == char:
				suggest = []prompt.Suggest{
					{Text: "-1", Description: "list one file per line"},
					{Text: "-a", Description: "do not ignore entries starting with"},
					{Text: "-f", Description: "do not sort"},
					{Text: "-h", Description: "with -l, print sizes like 1K 234M 2G etc."},
					{Text: "-l", Description: "use a long listing format"},
					{Text: "-n", Description: "list numeric user and group IDs"},
					{Text: "-r", Description: "reverse order while sorting"},
					{Text: "-S", Description: "sort by file size, largest first"},
					{Text: "-t", Description: "sort by modification time, newest first"},
				}
				return prompt.FilterHasPrefix(suggest, d.GetWordBeforeCursor(), false)

			default:
				return sc.PathComplete(false, 1, d)
			}
		case "lmkdir":
			switch {
			case "-" == char:
				suggest = []prompt.Suggest{
					{Text: "-p", Description: "no error if existing, make parent directories as needed"},
				}
				return prompt.FilterHasPrefix(suggest, d.GetWordBeforeCursor(), false)

			default:
				return sc.PathComplete(false, 1, d)
			}

		// case "ln":
		case "lpwd":
		case "ls":
			// switch options or path
			switch {
			case "-" == char:
				suggest = []prompt.Suggest{
					{Text: "-1", Description: "list one file per line"},
					{Text: "-a", Description: "do not ignore entries starting with"},
					{Text: "-f", Description: "do not sort"},
					{Text: "-h", Description: "with -l, print sizes like 1K 234M 2G etc."},
					{Text: "-l", Description: "use a long listing format"},
					{Text: "-n", Description: "list numeric user and group IDs"},
					{Text: "-r", Description: "reverse order while sorting"},
					{Text: "-S", Description: "sort by file size, largest first"},
					{Text: "-t", Description: "sort by modification time, newest first"},
				}
				return prompt.FilterHasPrefix(suggest, d.GetWordBeforeCursor(), false)

			default:
				return sc.PathComplete(true, 1, d)
			}

		// case "lumask":
		case "mkdir":
			switch {
			case "-" == char:
				suggest = []prompt.Suggest{
					{Text: "-p", Description: "no error if existing, make parent directories as needed"},
				}

			default:
				return sc.PathComplete(true, 1, d)
			}

		case "put":
			switch {
			case strings.Count(d.CurrentLineBeforeCursor(), " ") == 1: // local
				return sc.PathComplete(false, 1, d)
			case strings.Count(d.CurrentLineBeforeCursor(), " ") == 2: // remote
				return sc.PathComplete(true, 2, d)
			}
		case "pwd":
		case "quit":
		case "rename":
			return sc.PathComplete(true, 1, d)
		case "rm":
			return sc.PathComplete(true, 1, d)
		case "rmdir":
			return sc.PathComplete(true, 1, d)
		case "symlink":
			//
		default:
		}
	}
	return prompt.FilterHasPrefix(suggest, d.GetWordBeforeCursor(), true)
}

//
func (sc *sftpClient) PathComplete(remote bool, num int, d prompt.Document) []prompt.Suggest {
	// suggest
	var suggest []prompt.Suggest

	// Get cursor left
	left := d.CurrentLineBeforeCursor()

	// Get cursor char(string)
	char := ""
	if len(left) > 0 {
		char = string(left[len(left)-1])
	}

	// get last slash place
	word := d.GetWordBeforeCursor()
	sp := strings.LastIndex(word, "/")
	if len(word) > 0 {
		word = word[sp+1:]
	}

	switch remote {
	case true:
		// update sc.RemoteComplete
		switch {
		case "/" == char: // char is slach or
			sc.getRemoteComplete(d.GetWordBeforeCursor())
		case " " == char && strings.Count(d.CurrentLineBeforeCursor(), " ") == num:
			sc.getRemoteComplete(d.GetWordBeforeCursor())
		}
		suggest = sc.rComplete

	case false:
		// update sc.RemoteComplete
		switch {
		case "/" == char: // char is slach or
			sc.getLocalComplete(d.GetWordBeforeCursor())
		case " " == char && strings.Count(d.CurrentLineBeforeCursor(), " ") == num:
			sc.getLocalComplete(d.GetWordBeforeCursor())
		}
		suggest = sc.lComplete

	}

	return prompt.FilterHasPrefix(suggest, "_kua_"+word, false)
}

//
func (sc *sftpClient) getRemoteComplete(fp string) {
	// create suggest slice
	var suggest []prompt.Suggest

	// set rpath
	var rpath string
	if path.IsAbs(fp) {
		rpath = fp
	} else {
		rpath = path.Join(sc.rWorkDir, fp)
	}
	// check rpath
	stat, err := sc.client.Stat(rpath)
	if err != nil {
		return
	}
	if stat.IsDir() {
		rpath = rpath + "/*"
	} else {
		rpath = rpath + "*"
	}

	// get path list
	globlist, err := sc.client.Glob(rpath)
	if err != nil {
		return
	}

	// create suggest
	for _, p := range globlist {
		// create suggest
		pinfo, err := sc.client.Stat(p)
		if err != nil {
			continue
		}
		var desc string
		if pinfo.IsDir() {
			desc = "DIR"
		} else {
			desc = "File"
		}
		sug := prompt.Suggest{
			Text:        path.Base(p),
			Description: desc,
		}
		// append ps.Complete
		suggest = append(suggest, sug)
	}

	// sort
	sort.SliceStable(suggest, func(i, j int) bool { return suggest[i].Text < suggest[j].Text })
	// set suggest to struct
	sc.rComplete = suggest
}

//
func (sc *sftpClient) getLocalComplete(fp string) {
	// create suggest slice
	var suggest []prompt.Suggest
	stat, err := os.Lstat(fp)
	if err != nil {
		return
	}

	// dir check
	var lpath string
	if stat.IsDir() {
		lpath = fp + "/*"
	} else {
		lpath = fp + "*"
	}

	// get globlist
	globlist, err := filepath.Glob(lpath)
	if err != nil {
		return
	}

	// set path
	for _, lp := range globlist {
		lp = filepath.Base(lp)
		sug := prompt.Suggest{
			Text: lp,
			//Description: "local path.",
		}

		suggest = append(suggest, sug)
	}

	sc.lComplete = suggest
}

type bar struct {
	pb *pb.ProgressBar
}

// 创建PB
func (b *bar) NewBar(name string, total int64) {
	if b.pb == nil {
		b.pb = new(pb.ProgressBar)
		b.pb.SetWriter(os.Stdout)
		b.pb.Set(pb.Bytes, true)
		b.pb.SetWidth(100)
	}
	b.pb.SetTotal(total)
	b.pb.SetCurrent(0)
	// todo 设置输出字符
	//tmpl := `{{ counters . "%s/%s" "%s/?"}} {{ bar . "[" "=" (cycle . ">" ) "." "]"}} {{speed . | rndcolor }} {{percent .}} {{rtime . "%s" "%s" "???"}} {{string . "my_green_string" | green}} {{string . "my_blue_string" | blue}}`
	tmpl := name + ` {{ counters . "%s/%s" "%s/?"}} {{ bar . "[" "=" (cycle . ">" ) "." "]"}} {{speed .}} {{percent .}} {{rtime . "%s" "%s" "???"}}`
	b.pb.SetTemplateString(tmpl)
}

func (b *bar) Write(p []byte) (int, error) {
	n := len(p)
	b.pb.Add(n)
	return n, nil
}

func (b *bar) Read(p []byte) (int, error) {
	n := len(p)
	b.pb.Add(n)
	return n, nil
}
