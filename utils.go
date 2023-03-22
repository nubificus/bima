package main

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"time"
)

type dockerdata struct {
	unikernel string
	extrafile string
	utype     string
	cmdline   string
}

func (d dockerdata) createFile(dir string) error {
	docker := "FROM scratch\n"
	if d.utype != "" {
		docker += "LABEL \"com.urunc.unikernel.type\"=\"" + encode(d.utype) + "\"\n"
	}
	if d.cmdline != "" {
		docker += "LABEL \"com.urunc.unikernel.cmdline\"=\"" + encode(d.cmdline) + "\"\n"
	}
	docker += "LABEL \"com.urunc.unikernel.binary\"=\"" + encode("/unikernel/"+d.unikernel) + "\"\n"
	docker += "COPY " + d.unikernel + " /unikernel/\n"
	docker += "COPY " + "urunc.json" + " /\n"
	if d.extrafile != "" {
		docker += "COPY " + d.extrafile + " /extra/\n"
	}
	data := []byte(docker)

	err := os.WriteFile(dir+"Dockerfile", data, 0644)
	return err
}

func handle_err(err error) {
	fmt.Println("ERR: " + err.Error())
	os.Exit(1)
}

func commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}

func ensureDependencies() {
	if !commandExists("docker") {
		handle_err(errors.New("docker is not installed"))
	}
	if !commandExists("ctr") {
		handle_err(errors.New("ctr is not installed"))
	}
}

func fileExists(filename string) (bool, error) {
	file, err := os.Open(filename)
	if err != nil {
		return false, err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return false, err
	}
	if fileInfo.IsDir() {
		return false, nil
	}
	return true, nil
}

func randomTemp() string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	rand.Seed(time.Now().UnixNano())
	b := make([]rune, 10)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func copyFile(source string, target string) error {
	fin, err := os.Open(source)
	if err != nil {
		return err
	}
	defer fin.Close()

	fout, err := os.Create(target)
	if err != nil {
		return err
	}
	defer fout.Close()

	_, err = io.Copy(fout, fin)

	if err != nil {
		return err
	}
	return nil
}

func encode(data string) string {
	encoded := base64.StdEncoding.EncodeToString([]byte(data))
	return encoded
}

func dirExists(filename string) (bool, error) {
	file, err := os.Open(filename)
	if err != nil {
		return false, err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return false, err
	}
	if !fileInfo.IsDir() {
		return false, nil
	}
	return true, nil
}

func executeFromString(cmdline string) error {
	args := strings.Split(cmdline, " ")
	bin, args := args[0], args[1:]
	cmd := exec.Command(bin, args...)
	// return cmd.Run()

	outmsg, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(string(outmsg))
		return err
	}
	return nil
}
