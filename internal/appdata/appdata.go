package appdata

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const APPDATA_PATH = "/.local/share/tssc"
const APPDATA_FILE_NAME = "data.json"

type Appdata struct {
	Urls map[string]string `json:"urls"`
}

var AppData *Appdata

func Load() {
	appdataFilePath := getAppdataFilePath()

	fileContent := ensureReadFile(appdataFilePath)
	d := &Appdata{}
	if err := json.Unmarshal(fileContent, d); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	AppData = d
}

func Save(data *Appdata) {
	raw, err := json.Marshal(data)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	path := getAppdataFilePath()
	if err := os.WriteFile(path, raw, 0700); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func ensureReadFile(path string) (contents []byte) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			createAppdataFileOrPanic()
			return ensureReadFile(path)
		}

		fmt.Println(err.Error())
		os.Exit(1)
	}

	return data
}

func createAppdataFileOrPanic() {
	filePath := getAppdataFilePath()

	if _, err := os.Stat(filePath); err != nil {
		if os.IsNotExist(err) {
			homedir := getHomedirOrPanic()
			appdataPath := filepath.Join(homedir, APPDATA_PATH)
			err := os.MkdirAll(appdataPath, 0700)
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}
		} else {
			fmt.Println(err.Error())
			os.Exit(1)
		}
	}

	file, err := os.Create(filePath)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	defaultAppdata := &Appdata{
		Urls: map[string]string{},
	}

	Save(defaultAppdata)

	if err := file.Close(); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func getAppdataFilePath() string {
	homedir := getHomedirOrPanic()
	path := filepath.Join(homedir, APPDATA_PATH)
	appdataFilePath := filepath.Join(path, APPDATA_FILE_NAME)

	return appdataFilePath
}

func getHomedirOrPanic() string {
	homedir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	return homedir
}
