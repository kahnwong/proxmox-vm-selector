package main

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/luthermonson/go-proxmox"
)

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
