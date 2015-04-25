package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

//------------------------------------------------------------------------------------------------------------------------------
// convenience

func ifExit(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func home() string {
	usr, err := user.Current()
	ifExit(err)
	return usr.HomeDir
}

//------------------------------------------------------------------------------------------------------------------------------
// keys

// TODO: once ssh connections are established,
// overwrite the priv keys memory buffers
func decryptKeyFile(keyPath string) []byte {
	buf := new(bytes.Buffer)
	cmd := exec.Command("openssl", "rsa", "-in", keyPath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = buf
	cmd.Stderr = os.Stderr
	ifExit(cmd.Run())

	privateKey := buf.Bytes()
	return privateKey
}

func makeClientConfig(userName string, privateKey []byte) *ssh.ClientConfig {
	signer, err := ssh.ParsePrivateKey(privateKey)
	ifExit(err)
	return &ssh.ClientConfig{
		User: userName,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
	}
}

//------------------------------------------------------------------------------------------------------------------------------
// time/logging

const (
	layoutDir  = "2006-01-02"
	layoutFile = "15_04_MST"
)

// splits time into date and time (seconds precision)
// dir=date, file=time
func timeToDirFile(t time.Time) (string, string) {
	dir := t.Format(layoutDir)
	file := t.Format(layoutFile)
	dir = path.Join(rootDir, dir)
	if _, err := os.Stat(dir); err != nil {
		ifExit(os.MkdirAll(dir, 0700))
	}
	return dir, file
}

//------------------------------------------------------------------------------------------------------------------------------
// hosts

// return a single host or a list from file
func loadHosts(host string) []string {
	if _, err := os.Stat(host); err == nil {
		b, err := ioutil.ReadFile(host)
		ifExit(err)
		return strings.Split(string(b), "\n")
	} else {
		return []string{host}
	}
}
