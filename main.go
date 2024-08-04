package main

import (
	"context"
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/luthermonson/go-proxmox"
	"github.com/rs/zerolog/log"
)

var (
	virtualMachinesToPowerOn []string // huh data model
)

func getVirtualMachinesToPowerOff(vms proxmox.VirtualMachines, virtualMachinesToPowerOn []string) []string {
	var virtualMachinesAll []string
	for _, v := range vms {
		virtualMachinesAll = append(virtualMachinesAll, v.Name)
	}

	return subtractArrays(virtualMachinesAll, virtualMachinesToPowerOn)
}

func startVm(vms proxmox.VirtualMachines, virtualMachinesToPowerOn []string, taskChan chan error) {
	vmPowerOnCounter := 0
	defer close(taskChan)
	for _, vmName := range virtualMachinesToPowerOn {
		for _, vm := range vms {
			if vmName == vm.Name {
				if vm.Status != "running" {
					fmt.Printf("▶️ Starting %s...\n", vmName)
					vmPowerOnCounter = vmPowerOnCounter + 1

					_, err := vm.Start(context.Background())
					taskChan <- err
				}
			}
		}
	}

	if vmPowerOnCounter == 0 {
		fmt.Println("No virtual machines to start")
	}
}

func stopVm(vms proxmox.VirtualMachines, virtualMachinesToPowerOff []string, taskChan chan error) {
	vmPowerOffCounter := 0
	defer close(taskChan)
	for _, vmName := range virtualMachinesToPowerOff {
		for _, vm := range vms {
			if vmName == vm.Name {
				if vm.Status == "running" {
					fmt.Printf("⏸️ Stopping %s...\n", vmName)
					vmPowerOffCounter = vmPowerOffCounter + 1

					_, err := vm.Stop(context.Background())
					taskChan <- err
				}
			}
		}
	}
	if vmPowerOffCounter == 0 {
		fmt.Println("No virtual machines to stop")
	}
}

func main() {
	// init
	client := initClient()

	// print version
	version, err := getVersion(client)
	if err != nil {
		log.Fatal().Err(err).Msg("Cannot obtain Proxmox VE Version.")
	} else {
		fmt.Printf("Proxmox VE Version: %s\n", version)
	}

	// list VMs
	node, err := getNode(client)
	if err != nil {
		log.Fatal().Err(err).Msgf("Specified Proxmox VE Node does not exist.")
	}

	vms, err := getVMs(node)
	if err != nil {
		log.Fatal().Err(err).Msgf("Cannot obtain list of VMs.")
	}

	// init form
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
		log.Fatal().Err(err)
	}

	// start/stop VMs
	virtualMachinesToPowerOff := getVirtualMachinesToPowerOff(vms, virtualMachinesToPowerOn)

	startVmChan := make(chan error)
	go startVm(vms, virtualMachinesToPowerOn, startVmChan)
	for err := range startVmChan {
		if err != nil {
			log.Error().Msg("Starting VM Failed")
		}
	}

	stopVmChan := make(chan error)
	go stopVm(vms, virtualMachinesToPowerOff, stopVmChan)
	for err := range stopVmChan {
		if err != nil {
			log.Err(err).Msg("Stopping VM Failed")
		}
	}
}
