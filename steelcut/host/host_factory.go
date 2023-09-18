package host

import (
	"context"
	"fmt"

	"github.com/steelcutops/steelcut/steelcut/commandmanager"
	"github.com/steelcutops/steelcut/steelcut/filemanager"
	"github.com/steelcutops/steelcut/steelcut/hostmanager"
	"github.com/steelcutops/steelcut/steelcut/networkmanager"
)

func NewHost(hostname string) (HostInterface, error) {
	var ch ConcreteHost
	osType, err := ch.DetermineOS(context.TODO())
	if err != nil {
		return nil, err
	}

	switch osType {
	case LinuxUbuntu, LinuxDebian, LinuxFedora, LinuxRedHat, LinuxCentOS, LinuxArch, LinuxOpenSUSE:
		ch = configureLinuxHost(hostname)
	case Darwin:
		ch = configureMacHost(hostname)
	default:
		return nil, fmt.Errorf("unsupported operating system: %s", osType)
	}

	return &ch, nil
}

func configureLinuxHost(hostname string) ConcreteHost {
	cmdManager := &commandmanager.UnixCommandManager{Hostname: hostname}

	return ConcreteHost{
		CommandManager: cmdManager,
		FileManager:    &filemanager.UnixFileManager{CommandManager: cmdManager},
		HostManager:    &hostmanager.UnixHostManager{CommandManager: cmdManager},
		NetworkManager: &networkmanager.UnixNetworkManager{CommandManager: cmdManager},
		ServiceManager: &LinuxServiceManager{},
		PackageManager: &LinuxPackageManager{},
	}
}

func configureMacHost(hostname string) ConcreteHost {
	cmdManager := &commandmanager.UnixCommandManager{Hostname: hostname}

	return ConcreteHost{
		CommandManager: cmdManager,
		FileManager:    &filemanager.UnixFileManager{CommandManager: cmdManager},
		HostManager:    &hostmanager.UnixHostManager{CommandManager: cmdManager},
		NetworkManager: &networkmanager.UnixNetworkManager{CommandManager: cmdManager},
		ServiceManager: &DarwinServiceManager{},
		PackageManager: &DarwinPackageManager{},
	}
}
