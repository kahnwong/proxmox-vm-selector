package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/luthermonson/go-proxmox"

	"github.com/charmbracelet/huh"
	"github.com/joho/godotenv"
)

// main
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

func getVirtualMachinesToPowerOff(vms proxmox.VirtualMachines, virtualMachinesToPowerOn []string) []string {
	var virtualMachinesAll []string
	for _, v := range vms {
		virtualMachinesAll = append(virtualMachinesAll, v.Name)
	}

	return subtractArrays(virtualMachinesAll, virtualMachinesToPowerOn)
}

func startVm(vms proxmox.VirtualMachines, virtualMachinesToPowerOn []string) {
	vmPowerOnCounter := 0
	for _, vmName := range virtualMachinesToPowerOn {
		for _, vm := range vms {
			if vmName == vm.Name {
				if vm.Status != "running" {
					fmt.Printf("▶️ Starting %s...\n", vmName)
					vmPowerOnCounter = vmPowerOnCounter + 1

					_, err := vm.Start(context.Background())
					if err != nil {
						fmt.Println(err)
					}
				}
			}
		}
	}
	if vmPowerOnCounter == 0 {
		fmt.Println("No virtual machines to start")
	}
}

func stopVm(vms proxmox.VirtualMachines, virtualMachinesToPowerOff []string) {
	for _, vmName := range virtualMachinesToPowerOff {
		for _, vm := range vms {
			if vmName == vm.Name {
				fmt.Printf("⏸️ Stopping %s...\n", vmName)
				_, err := vm.Stop(context.Background())
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

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Loading env from env var instead...")
	}

	// init
	client := initClient()

	// print version
	version, err := getVersion(client)
	if err != nil {
		slog.Error("Cannot obtain Proxmox VE Version.")
	}
	fmt.Printf("Proxmox VE Version: %s\n", version)

	// list VMs
	node, err := getNode(client)
	if err != nil {
		slog.Error(fmt.Sprintf("Specified Proxmox VE Node `%s` does not exist.", os.Getenv("PROXMOX_VE_NODENAME")))
	}

	vms, err := getVMs(node)
	if err != nil {
		fmt.Println(err)
	}

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
	virtualMachinesToPowerOff := getVirtualMachinesToPowerOff(vms, virtualMachinesToPowerOn)
	startVm(vms, virtualMachinesToPowerOn)
	stopVm(vms, virtualMachinesToPowerOff)
}
