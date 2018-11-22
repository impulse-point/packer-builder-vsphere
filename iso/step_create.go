package iso

import (
	"context"
	"fmt"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/jetbrains-infra/packer-builder-vsphere/common"
	"github.com/jetbrains-infra/packer-builder-vsphere/driver"
)

type Network struct {
	Network       string `mapstructure:"network"`
	NetworkCard   string `mapstructure:"network_card"`
}

type CreateConfig struct {
	Version     uint   `mapstructure:"vm_version"`
	GuestOSType string `mapstructure:"guest_os_type"`
	Firmware    string `mapstructure:"firmware"`

	DiskControllerType  string `mapstructure:"disk_controller_type"`
	DiskSize            int64  `mapstructure:"disk_size"`
	DiskThinProvisioned bool   `mapstructure:"disk_thin_provisioned"`

	Network       string `mapstructure:"network"`
	NetworkCard   string `mapstructure:"network_card"`

	Networks []Network `mapstructure:"networks"`

	USBController bool   `mapstructure:"usb_controller"`
}

func ToNetworkConfig(o []Network) []driver.NetworkConfig {
	var net []driver.NetworkConfig

	if o != nil {
		for _, item := range o {
			n := driver.NetworkConfig{
				Network: item.Network,
				NetworkCard: item.NetworkCard,
			}
			net = append(net, n)
		}
	}
	return net
}

func (c *CreateConfig) Prepare() []error {
	var errs []error

	if c.DiskSize == 0 {
		errs = append(errs, fmt.Errorf("'disk_size' is required"))
	}

	if c.GuestOSType == "" {
		c.GuestOSType = "otherGuest"
	}

	if c.Firmware != "" && c.Firmware != "bios" && c.Firmware != "efi" {
		errs = append(errs, fmt.Errorf("'firmware' must be 'bios' or 'efi'"))
	}

	return errs
}

type StepCreateVM struct {
	Config   *CreateConfig
	Location *common.LocationConfig
}

func (s *StepCreateVM) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	d := state.Get("driver").(*driver.Driver)

	ui.Say("Creating VM...")
	vm, err := d.CreateVM(&driver.CreateConfig{
		DiskThinProvisioned: s.Config.DiskThinProvisioned,
		DiskControllerType:  s.Config.DiskControllerType,
		DiskSize:            s.Config.DiskSize,
		Annotation:          "",
		Name:                s.Location.VMName,
		Folder:              s.Location.Folder,
		Cluster:             s.Location.Cluster,
		Host:                s.Location.Host,
		ResourcePool:        s.Location.ResourcePool,
		Datastore:           s.Location.Datastore,
		GuestOS:             s.Config.GuestOSType,
		Network:             s.Config.Network,
		NetworkCard:         s.Config.NetworkCard,

		Networks:            ToNetworkConfig(s.Config.Networks),
		USBController:       s.Config.USBController,
		Version:             s.Config.Version,
		Firmware:            s.Config.Firmware,
	})
	if err != nil {
		state.Put("error", fmt.Errorf("error creating vm: %v", err))
		return multistep.ActionHalt
	}
	state.Put("vm", vm)

	return multistep.ActionContinue
}

func (s *StepCreateVM) Cleanup(state multistep.StateBag) {
	_, cancelled := state.GetOk(multistep.StateCancelled)
	_, halted := state.GetOk(multistep.StateHalted)
	if !cancelled && !halted {
		return
	}

	ui := state.Get("ui").(packer.Ui)

	st := state.Get("vm")
	if st == nil {
		return
	}
	vm := st.(*driver.VirtualMachine)

	ui.Say("Destroying VM...")
	err := vm.Destroy()
	if err != nil {
		ui.Error(err.Error())
	}
}
