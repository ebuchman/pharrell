package main

import (
	"bytes"
	"flag"
	"fmt"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path"
	"strings"
	"time"
)

var (
	rootDir = path.Join(home(), ".pharrell")

	userF = flag.String("u", "root", "username")
	hostF = flag.String("h", "", "host")
)

func home() string {
	usr, err := user.Current()
	ifExit(err)
	return usr.HomeDir
}

func init() {
	if _, err := os.Stat(rootDir); err != nil {
		ifExit(os.MkdirAll(rootDir, 0700))
	}
}

func main() {
	flag.Parse()

	if *hostF == "" {
		fmt.Println("Host flag is mandatory.")
		os.Exit(1)
	}

	if len(flag.Args()) < 2 {
		fmt.Println("Enter a source and destination")
		os.Exit(1)
	}

	userName := *userF

	src, dst := flag.Args()[0], flag.Args()[1]
	_, err := os.Stat(src)
	ifExit(err)

	// if host its a file, parse
	var hosts []string
	if _, err := os.Stat(*hostF); err == nil {
		b, err := ioutil.ReadFile(*hostF)
		ifExit(err)
		hosts = strings.Split(string(b), "\n")
	} else {
		hosts = []string{*hostF}
	}

	buf := new(bytes.Buffer)
	cmd := exec.Command("openssl", "rsa", "-in", path.Join(home(), ".ssh/id_rsa"))
	cmd.Stdin = os.Stdin
	cmd.Stdout = buf
	cmd.Stderr = os.Stderr
	ifExit(cmd.Run())

	privateKey := buf.Bytes()

	signer, err := ssh.ParsePrivateKey(privateKey)
	ifExit(err)
	clientConfig := &ssh.ClientConfig{
		User: userName,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
	}

	done := make(chan error)
	liftoff := time.Now()
	for _, host := range hosts {
		go copyToHost(clientConfig, src, dst, host, done)
	}

	const (
		layoutDir  = "2006-01-02"
		layoutFile = "15_04_MST"
	)

	dir := liftoff.Format(layoutDir)
	file := liftoff.Format(layoutFile)
	dir = path.Join(rootDir, dir)
	if _, err := os.Stat(dir); err != nil {
		ifExit(os.MkdirAll(dir, 0700))
	}
	for _, host := range hosts {
		b := <-done
		spl := strings.Split(host, ":")
		addr := spl[0]

		if b != nil {
			err := ioutil.WriteFile(path.Join(dir, file+"_"+addr), []byte(b.Error()), 0600)
			if err != nil {
				fmt.Println("Error writing host to file: ", host, err)
			}
		}
	}
}

func copyToHost(config *ssh.ClientConfig, src, dst, host string, done chan error) {
	client, err := ssh.Dial("tcp", host, config)
	if err != nil {
		done <- fmt.Errorf("Failed to dial: %s %s", host, err.Error())
		return
	}
	scp := NewScp(client)

	f, _ := os.Stat(src)
	if f.IsDir() {
		err = scp.PushDir(src, dst)
	} else {
		err = scp.PushFile(src, dst)
	}

	done <- err
}

func ifExit(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
