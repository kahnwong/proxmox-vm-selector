package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"sort"
	"unicode"

	"github.com/luthermonson/go-proxmox"
)

// base
func initClient() *proxmox.Client {

	credentials := proxmox.Credentials{
		Username: config.ProxmoxVEUsername,
		Password: config.ProxmoxVEPassword,
	}
	proxmoxAPIEndpoint := fmt.Sprintf("https://%s/api2/json", config.ProxmoxVEHost)

	var client *proxmox.Client
	isIP := unicode.IsDigit(rune(config.ProxmoxVEHost[0]))
	if !isIP {
		client = proxmox.NewClient(
			proxmoxAPIEndpoint,
			proxmox.WithCredentials(&credentials),
		)
	} else {
		insecureHTTPClient := http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			},
		}
		client = proxmox.NewClient(
			proxmoxAPIEndpoint,
			proxmox.WithCredentials(&credentials),
			proxmox.WithHTTPClient(&insecureHTTPClient),
		)
	}

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
	} else {
		// sort output by VM name
		sort.Slice(vms, func(i, j int) bool {
			return vms[i].Name < vms[j].Name
		})
		return vms, nil
	}
}
