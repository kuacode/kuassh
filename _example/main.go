package main

import (
	"bufio"
	"github.com/mattn/go-tty"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
	"log"
	"os"
	"time"
	"unicode/utf8"
)

func main() {
	check := func(err error, msg string) {
		if err != nil {
			log.Fatalf("%s error: %v", msg, err)
		}
	}

	client, err := ssh.Dial("tcp", "172.17.111.16:2222", &ssh.ClientConfig{
		User: "kcode",
		Auth: []ssh.AuthMethod{ssh.Password("123456")},
		//需要验证服务端，不做验证返回nil就可以，点击HostKeyCallback看源码就知道了
		// HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		// 	return nil
		// },
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	})
	check(err, "dial")

	session, err := client.NewSession()
	check(err, "new session")
	defer session.Close()
	//
	t, err := tty.Open()
	if err != nil {
		log.Fatal(err)
	}
	//fd := int(t.Input().Fd())
	//state, err := terminal.MakeRaw(fd)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//defer terminal.Restore(fd, state)

	//ofd := int(t.Output().Fd())
	ofd := int(os.Stdout.Fd())
	w, h, err := terminal.GetSize(ofd)
	ws, hs, err := t.Size()
	log.Println(ws, hs)
	if err != nil {
		log.Fatal(err)
	}
	//session.Stdout = os.Stdout
	//session.Stderr = os.Stderr
	//session.Stdin = os.Stdin
	session.Stdout = t.Output()
	session.Stderr = t.Output()
	stdinPipe, _ := session.StdinPipe()

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,     // 禁用回显（0禁用，1启动）
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, //output speed = 14.4kbaud
	}
	err = session.RequestPty("xterm", h, w, modes)
	check(err, "request pty")

	err = session.Shell()
	check(err, "start shell")

	go func() {
		for ws := range t.SIGWINCH() {
			session.WindowChange(ws.H, ws.W)
		}
	}()

	//clean, err := t.Raw()
	//if err != nil {
	//	log.Fatal(err)
	//}
	//defer clean()
	go func() {
		buf := make([]byte, 128)
		//rs := []rune{}
		for {
			//n, err := t.Input().Read(bs)
			//
			//if err != nil {
			//	continue
			//}
			//stdinPipe.Write(bs[:n])
			//r, _ := t.ReadRune()
			//if r == 0 {
			//	continue
			//}
			//stdinPipe.Write([]byte(string(r)))

			//	session.Stdout.Write([]byte(fmt.Sprint(r)))
			//} else {
			//
			//}
			reader := bufio.NewReaderSize(t.Input(), 128)
			r, _, _ := reader.ReadRune()

			n := utf8.EncodeRune(buf[:], r)
			for reader.Buffered() > 0 && n < 128 {
				r, _, _ = reader.ReadRune()
				n += utf8.EncodeRune(buf[n:], r)
			}
			stdinPipe.Write(buf[:n])
		}
	}()

	// send keepalive
	go func() {
		for {
			time.Sleep(time.Second * 10)
			client.SendRequest("keepalive@openssh.com", false, nil)
		}
	}()

	err = session.Wait()
	check(err, "return")
}
