package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sort"

	"github.com/charmbracelet/huh"
	"github.com/joho/godotenv"
	"github.com/luthermonson/go-proxmox"
)

// base
func initClient() *proxmox.Client {
	credentials := proxmox.Credentials{
		Username: os.Getenv("PROXMOX_VE_USERNAME"),
		Password: os.Getenv("PROXMOX_VE_PASSWORD"),
	}

	proxmoxAPIEndpoint := fmt.Sprintf("https://%s/api2/json", os.Getenv("PROXMOX_VE_HOST"))
	client := proxmox.NewClient(proxmoxAPIEndpoint,
		proxmox.WithCredentials(&credentials),
	)

	return client
}

func getVersion(client *proxmox.Client) string {
	version, err := client.Version(context.Background())
	if err != nil {
		panic(err)
	}

	return version.Release
}

// vm
func getNode(client *proxmox.Client) *proxmox.Node {
	node, err := client.Node(context.Background(), os.Getenv("PROXMOX_VE_NODENAME"))
	if err != nil {
		fmt.Println(err)
	}

	return node
}

func getVMs(node *proxmox.Node) proxmox.VirtualMachines {
	vms, err := node.VirtualMachines(context.Background())
	if err != nil {
		fmt.Println(err)
	}

	// sort output by VM name
	sort.Slice(vms, func(i, j int) bool {
		return vms[i].Name < vms[j].Name
	})

	return vms
}

// form
func createFormOptions(vms proxmox.VirtualMachines) []huh.Option[string] {
	var options []huh.Option[string]

	for _, v := range vms {
		vmName := v.Name
		vmStatus := v.Status

		vmStatusSymbolMap := map[string]string{
			"running": "✅",
			"stopped": "🛑",
		}

		var vmStatysSymbol string
		if value, ok := vmStatusSymbolMap[vmStatus]; ok {
			vmStatysSymbol = value
			vmDisplayName := fmt.Sprintf("%s %s", vmStatysSymbol, vmName)

			if vmStatus == "running" {
				options = append(options, huh.NewOption(vmDisplayName, vmName).Selected(true))

			} else if vmStatus == "stopped" {
				options = append(options, huh.NewOption(vmDisplayName, vmName))
			}
		} else {
			fmt.Println("Value not found in the map.")
		}
	}

	return options
}

// main
// subtractArrays subtracts elements of array2 from array1
func subtractArrays(array1, array2 []string) []string {
	var result []string

	// Create a map to efficiently check if an element is in array2
	exists := make(map[string]bool)
	for _, elem := range array2 {
		exists[elem] = true
	}

	// Iterate through array1 and add elements not in array2 to the result
	for _, elem := range array1 {
		if !exists[elem] {
			result = append(result, elem)
		}
	}

	return result
}

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Loading env from env var instead...")
	}

	// init
	client := initClient()

	// print version
	version := getVersion(client)
	fmt.Printf("Proxmox VE Version: %s\n", version)

	// list VMs
	node := getNode(client)
	vms := getVMs(node)

	// init form
	var (
		virtualMachinesToPowerOn []string
	)

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Select Virtual Machines you want to Power On").
				Options(
					createFormOptions(vms)...,
				).
				Value(&virtualMachinesToPowerOn),
		),
	)
	err = form.Run()
	if err != nil {
		log.Fatal(err)
	}

	// start/stop VMs
	var virtualMachinesAll []string
	for _, v := range vms {
		virtualMachinesAll = append(virtualMachinesAll, v.Name)
	}

	virtualMachinesToPowerOff := subtractArrays(virtualMachinesAll, virtualMachinesToPowerOn)

	for _, vmName := range virtualMachinesToPowerOn {
		for _, vm := range vms {
			if vmName == vm.Name {
				fmt.Printf("▶️ Starting %s...\n", vmName)
				_, err = vm.Start(context.Background())
				if err != nil {
					fmt.Println(err)
				}
			}
		}
	}
	if len(virtualMachinesToPowerOn) == 0 {
		fmt.Println("No virtual machines to start")
	}

	for _, vmName := range virtualMachinesToPowerOff {
		for _, vm := range vms {
			if vmName == vm.Name {
				fmt.Printf("⏸️ Stopping %s...\n", vmName)
				_, err = vm.Stop(context.Background())
				if err != nil {
					fmt.Println(err)
				}
			}
		}
	}
	if len(virtualMachinesToPowerOff) == 0 {
		fmt.Println("No virtual machines to stop")
	}
}
