package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
)

const ShellToUse = "bash"

func Shellout(command string) (string, string, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := exec.Command(ShellToUse, "-c", command)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

func createFileWithPermissions(path, fileName string, perm os.FileMode) error {
	// Combine path and filename to get the full file path
	fullPath := filepath.Join(path, fileName)

	// Create the file
	file, err := os.Create(fullPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Change the permissions of the file
	if err := os.Chmod(fullPath, perm); err != nil {
		return err
	}

	return nil
}

func createFolder(path string, name string) error {
	if _, err := exec.Command("mkdir", path+"/"+name).Output(); err != nil {
		return err
	} else {
		return nil
	}
}

func createFile(path string, name string) error {
	err := createFileWithPermissions(path, name, os.FileMode(0755))
	if err != nil {
		return err
	}
	return nil
}

func renameFileOrDir(path string, name string, newName string) error {
	oldPathName := path + "/" + name
	newPathName := path + "/" + newName
	if _, err := exec.Command("mv", oldPathName, newPathName).Output(); err != nil {
		return err
	}
	return nil
}

func deleteFileOrDir(path string, name string) error {
	if _, err := exec.Command("rm", "-rf", path+"/"+name).Output(); err != nil {
		return err
	}
	return nil
}
