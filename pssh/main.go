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

	if len(flag.Args()) < 2 {
		fmt.Println("Enter a quote enclosed command or file containing commands, followed by a host, list of hosts, or file containing hosts")
		os.Exit(1)
	}
	command := flag.Args()[0]
	// if its a file, parse
	if _, err := os.Stat(command); err == nil {
		b, err := ioutil.ReadFile(command)
		ifExit(err)
		command = composeCommand(b)
	}
	hosts := flag.Args()[1:]
	// if its a file, parse
	if _, err := os.Stat(hosts[0]); err == nil {
		b, err := ioutil.ReadFile(hosts[0])
		ifExit(err)
		hosts = strings.Split(string(b), "\n")
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
		User: "minty",
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
	}
	done := make(chan *bytes.Buffer)
	liftoff := time.Now()
	for _, host := range hosts {
		go runCommandOnHost(clientConfig, command, host, done)
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
		err := ioutil.WriteFile(path.Join(dir, file+"_"+addr), b.Bytes(), 0600)
		if err != nil {
			fmt.Println("Error writing host to file: ", host, err)
		}
	}
}

func runCommandOnHost(config *ssh.ClientConfig, cmd, host string, done chan *bytes.Buffer) {
	client, err := ssh.Dial("tcp", host, config)
	if err != nil {
		fmt.Println("Failed to dial: " + host + " " + err.Error())
	}

	// Each ClientConn can support multiple interactive sessions,
	// represented by a Session.
	session, err := client.NewSession()
	if err != nil {
		fmt.Println("Failed to create session: " + err.Error())
	}
	defer session.Close()

	// Once a Session is created, you can execute a single command on
	// the remote side using the Run method.
	b := new(bytes.Buffer)
	session.Stdout = b
	if err := session.Run(cmd); err != nil {
		panic("Failed to run: " + err.Error())
	}
	done <- b
}

func ifExit(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
