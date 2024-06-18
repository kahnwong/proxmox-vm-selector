package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/getsops/sops/v3/decrypt"
	"gopkg.in/yaml.v3"
)

// config
var config = readConfig()

type Config struct {
	ProxmoxVEHost     string `yaml:"PROXMOX_VE_HOST"`
	ProxmoxVEUsername string `yaml:"PROXMOX_VE_USERNAME"`
	ProxmoxVEPassword string `yaml:"PROXMOX_VE_PASSWORD"`
	ProxmoxVENodeName string `yaml:"PROXMOX_VE_NODENAME"`
}

func readConfig() Config {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	filename := filepath.Join(homeDir, ".config", "proxmox", "proxmox.sops.yaml")

	// Check if the file exists
	_, err = os.Stat(filename)

	if os.IsNotExist(err) {
		fmt.Printf("File %s does not exist.\n", filename)
		os.Exit(1)
	}

	var config Config

	data, err := decrypt.File(filename, "yaml")
	if err != nil {
		fmt.Println(fmt.Printf("Failed to decrypt: %v", err))
	}

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}

	return config
}
