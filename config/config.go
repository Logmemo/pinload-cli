package config

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
)

type Config struct {
	GalleryPath string `json:"galleryPath"`
}

func GetDefaultDownloadPath() string {
	homeDir, userHomeDirErr := os.UserHomeDir()
	if userHomeDirErr != nil {
		log.Fatalln(userHomeDirErr)
	}
	downloadPath := filepath.Join(homeDir, "Downloads")
	return downloadPath
}

func GetConfigPath() string {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exePath := filepath.Dir(ex)
	configName := "config.json"
	configPath := filepath.Join(exePath, configName)

	_, statErr := os.Stat(configPath)
	if statErr == nil {

		var config Config
		file, readErr := os.ReadFile(configPath)
		if readErr != nil {
			log.Fatalf("Failed to read file: %s", readErr)
		}

		unmarshalErr := json.Unmarshal(file, &config)
		if unmarshalErr != nil {
			log.Fatalf("Failed to read file: %s", unmarshalErr)
		}

		return config.GalleryPath
	} else {
		return ""
	}
}

func ChangeConfigPath(path string) {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exePath := filepath.Dir(ex)
	configName := "config.json"
	configPath := filepath.Join(exePath, configName)

	_, err = os.Stat(configPath)
	if err == nil {
		_, err2 := os.Stat(path)
		if err2 != nil {
			log.Fatal(err)
		}
		data := Config{
			GalleryPath: path,
		}

		jsonData, err := json.Marshal(data)
		err3 := os.WriteFile("config.json", []byte(jsonData), 0744)
		if err3 != nil {
			log.Fatal(err)
		}
	}
}

func CheckConfig() {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exePath := filepath.Dir(ex)
	configName := "config.json"
	configPath := filepath.Join(exePath, configName)

	_, err = os.Stat(configPath)
	if err != nil {

		config, err2 := os.Create(configPath)
		if err2 != nil {
			log.Fatalf("Failed to create file: %s", err)
		}
		defer config.Close()

		data := Config{
			GalleryPath: GetDefaultDownloadPath(),
		}
		jsonData, err := json.Marshal(data)
		err3 := os.WriteFile(configPath, []byte(jsonData), 0744)
		if err3 != nil {
			log.Fatal(err)
		}
	}
}
