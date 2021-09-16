package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Setting struct {
	Language string `json:"language"`
}

func NewSetting() *Setting {
	setting := Setting{
		Language: "en",
	}
	return &setting
}

func settingPath() (string, error) {
	exePath, err := os.Executable()
	if err != nil {
		return "", err
	}
	path := filepath.Dir(exePath) + "/setting.json"
	return path, nil
}

func ReadSetting() (*Setting, error) {
	settingPath, err := settingPath()
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(settingPath); err != nil {
		newSetting := NewSetting()
		if err := WriteSetting(newSetting); err != nil {
			return nil, err
		}
		return newSetting, nil
	}

	bytes, err := os.ReadFile(settingPath)
	if err != nil {
		return nil, err
	}
	var setting *Setting
	if err := json.Unmarshal(bytes, &setting); err != nil {
		return nil, err
	}
	return setting, nil
}

func WriteSetting(setting *Setting) error {
	settingPath, err := settingPath()
	if err != nil {
		return err
	}

	file, err := os.OpenFile(settingPath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	bytes, err := json.Marshal(setting)
	if err != nil {
		return err
	}
	fmt.Fprintln(file, string(bytes))

	return nil
}
