package main

import (
	"log"
	"os"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
)

func main() {
	check := func(err error, msg string) {
		if err != nil {
			log.Fatalf("%s error: %v", msg, err)
		}
	}

	client, err := ssh.Dial("tcp", "149.28.25.177:22", &ssh.ClientConfig{
		User: "root",
		Auth: []ssh.AuthMethod{ssh.Password("%9aA-jR1[973FBn$")},
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
	fd := int(os.Stdin.Fd())
	state, err := terminal.MakeRaw(fd)
	if err != nil {
		log.Fatal(err)
	}
	defer terminal.Restore(fd, state)
	w, h, err := terminal.GetSize(fd)
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr
	session.Stdin = os.Stdin

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
		var (
			ow = w
			oh = h
		)
		for {
			cw, ch, err := terminal.GetSize(fd)
			if err != nil {
				break
			}

			if cw != ow || ch != oh {
				err = session.WindowChange(ch, cw)
				if err != nil {
					break
				}
				ow = cw
				oh = ch
			}
			time.Sleep(time.Second)
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
