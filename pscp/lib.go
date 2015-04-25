// code originally taken from https://github.com/gnicod/goscplib
// but modified to work

package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"golang.org/x/crypto/ssh"
)

//Constants
const (
	SCP_PUSH_BEGIN_FILE       = "C"
	SCP_PUSH_BEGIN_FOLDER     = "D"
	SCP_PUSH_BEGIN_END_FOLDER = "0"
	SCP_PUSH_END_FOLDER       = "E"
	SCP_PUSH_END              = "\x00"
)

type Scp struct {
	client *ssh.Client
}

func GetPerm(f *os.File) (perm string) {
	fileStat, _ := f.Stat()
	mod := fileStat.Mode()
	if mod > (1 << 9) {
		mod = mod % (1 << 9)
	}
	return fmt.Sprintf("%#o", uint32(mod))
}

//Initializer
func NewScp(clientConn *ssh.Client) *Scp {
	return &Scp{
		client: clientConn,
	}
}

//Push one file to server
func (scp *Scp) PushFile(src string, dest string) error {
	session, err := scp.client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()
	go func() {
		w, _ := session.StdinPipe()
		defer w.Close()
		fileSrc, srcErr := os.Open(src)
		defer fileSrc.Close()
		//fileStat, err := fileSrc.Stat()
		if srcErr != nil {
			log.Fatalln("Failed to open source file: " + srcErr.Error())
		}
		//Get file size
		srcStat, statErr := fileSrc.Stat()
		if statErr != nil {
			log.Fatalln("Failed to stat file: " + statErr.Error())
		}
		// According to https://blogs.oracle.com/janp/entry/how_the_scp_protocol_works
		// Print the file content
		fmt.Fprintln(w, SCP_PUSH_BEGIN_FILE+GetPerm(fileSrc), srcStat.Size(), filepath.Base(dest))
		io.Copy(w, fileSrc)
		fmt.Fprint(w, SCP_PUSH_END)
	}()
	if err := session.Run("/usr/bin/scp -rt " + filepath.Dir(dest)); err != nil {
		return err
	}
	return nil
}

//Push directory to server
func (scp *Scp) PushDir(src string, dest string) error {
	session, err := scp.client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()
	go func() {
		w, _ := session.StdinPipe()
		//w = os.Stdout
		defer w.Close()
		folderSrc, _ := os.Open(src)
		fmt.Fprintln(w, SCP_PUSH_BEGIN_FOLDER+GetPerm(folderSrc), SCP_PUSH_BEGIN_END_FOLDER, filepath.Base(dest))
		lsDir(w, src)
		fmt.Fprintln(w, SCP_PUSH_END_FOLDER)

	}()
	if err := session.Run("/usr/bin/scp -qrt " + dest); err != nil {
		return err
	}
	return nil
}

func prepareFile(w io.WriteCloser, src string) {
	fileSrc, srcErr := os.Open(src)
	defer fileSrc.Close()
	if srcErr != nil {
		log.Fatalln("Failed to open source file: " + srcErr.Error())
	}
	//Get file size
	srcStat, statErr := fileSrc.Stat()
	if statErr != nil {
		log.Fatalln("Failed to stat file: " + statErr.Error())
	}
	// Print the file content
	fmt.Fprintln(w, SCP_PUSH_BEGIN_FILE+GetPerm(fileSrc), srcStat.Size(), filepath.Base(src))
	io.Copy(w, fileSrc)
	fmt.Fprint(w, SCP_PUSH_END)
}

func lsDir(w io.WriteCloser, dir string) {
	fi, _ := ioutil.ReadDir(dir)
	//parcours des dossiers
	for _, f := range fi {
		if f.IsDir() {
			folderSrc, _ := os.Open(path.Join(dir, f.Name()))
			defer folderSrc.Close()
			fmt.Fprintln(w, SCP_PUSH_BEGIN_FOLDER+GetPerm(folderSrc), SCP_PUSH_BEGIN_END_FOLDER, f.Name())
			lsDir(w, path.Join(dir, f.Name()))
			fmt.Fprintln(w, SCP_PUSH_END_FOLDER)
		} else {
			prepareFile(w, path.Join(dir, f.Name()))
		}
	}
}
