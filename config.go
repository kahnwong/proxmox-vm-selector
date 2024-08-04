package main

import (
	"os"
	"path/filepath"

	"github.com/getsops/sops/v3/decrypt"
	"github.com/rs/zerolog/log"
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
		log.Fatal().Err(err).Msg("Error reading user home directory")
	}
	filename := filepath.Join(homeDir, ".config", "proxmox", "proxmox.sops.yaml")

	// Check if the file exists
	_, err = os.Stat(filename)

	if os.IsNotExist(err) {
		log.Fatal().Err(err).Msgf("File %s does not exist.\n", filename)
	}

	var config Config

	data, err := decrypt.File(filename, "yaml")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to decrypt config file")
	}

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to unmarshal config file")
	}

	return config
}
