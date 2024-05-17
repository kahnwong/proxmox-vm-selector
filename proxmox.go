package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"

	"github.com/getsops/sops/v3/decrypt"
	"github.com/luthermonson/go-proxmox"
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

// base
func initClient() *proxmox.Client {
	credentials := proxmox.Credentials{
		Username: config.ProxmoxVEUsername,
		Password: config.ProxmoxVEPassword,
	}

	proxmoxAPIEndpoint := fmt.Sprintf("https://%s/api2/json", config.ProxmoxVEHost)
	client := proxmox.NewClient(proxmoxAPIEndpoint,
		proxmox.WithCredentials(&credentials),
	)

	return client
}

func getVersion(client *proxmox.Client) (string, error) {
	version, err := client.Version(context.Background())

	return version.Release, err
}

// vm
func getNode(client *proxmox.Client) (*proxmox.Node, error) {
	node, err := client.Node(context.Background(), config.ProxmoxVENodeName)

	return node, err
}

func getVMs(node *proxmox.Node) (proxmox.VirtualMachines, error) {
	vms, err := node.VirtualMachines(context.Background())
	if err != nil {
		return nil, err
	}

	// sort output by VM name
	sort.Slice(vms, func(i, j int) bool {
		return vms[i].Name < vms[j].Name
	})

	return vms, nil
}
