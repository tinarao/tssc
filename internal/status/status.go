package status

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

const LOCKFILE_DIR = "/etc/tssc/"
const LOCKFILE_NAME = "tssc.lock"

func Lock(alias string) error {
	_, err := os.ReadFile(getLockFilePath())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			createLockFile()
			return Lock(alias)
		}

		fmt.Println(err.Error())
		os.Exit(1)
	}

	if err := os.WriteFile(getLockFilePath(), []byte(alias), 0700); err != nil {
		return err
	}

	return nil
}

func Unlock() error {
	return os.Remove(getLockFilePath())
}

func IsLocked() bool {
	_, err := os.ReadFile(getLockFilePath())
	return err == nil
}

func getLockFilePath() string {
	return filepath.Join(LOCKFILE_DIR, LOCKFILE_NAME)
}

func createLockFile() {
	_, err := os.Stat(LOCKFILE_DIR)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			if err := os.Mkdir(LOCKFILE_DIR, 0700); err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}
		}
	}

	f, err := os.Create(getLockFilePath())
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Println("created")

	f.Close()
}
