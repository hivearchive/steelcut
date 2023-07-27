package steelcut

import (
	"fmt"
	"log"
	"os/exec"
	"os/user"

	"strings"

	"golang.org/x/crypto/ssh"
)

type Update struct {
	// Fields for the Update struct
}

type Host interface {
	CheckUpdates() ([]Update, error)
	RunCommand(cmd string) (string, error)
	ListPackages() ([]string, error)
	AddPackage(pkg string) error
	RemovePackage(pkg string) error
	UpgradePackage(pkg string) error
}

type UnixHost struct {
	Hostname      string
	User          string
	Password      string
	KeyPassphrase string
	OS            string
}

type MacOSHost struct {
	UnixHost
	PackageManager PackageManager
}

type LinuxHost struct {
	UnixHost
	PackageManager PackageManager
}

func (h LinuxHost) ListPackages() ([]string, error) {
	return h.PackageManager.ListPackages(h.UnixHost)
}

func (h LinuxHost) AddPackage(pkg string) error {
	return h.PackageManager.AddPackage(h.UnixHost, pkg)
}

func (h LinuxHost) RemovePackage(pkg string) error {
	return h.PackageManager.RemovePackage(h.UnixHost, pkg)
}

func (h LinuxHost) UpgradePackage(pkg string) error {
	return h.PackageManager.UpgradePackage(h.UnixHost, pkg)
}

func (h LinuxHost) CheckUpdates() ([]Update, error) {
	// Implement the update check for Linux hosts.
	return []Update{}, nil
}

func (h LinuxHost) RunCommand(cmd string) (string, error) {
	return h.UnixHost.RunCommand(cmd)
}

func (h MacOSHost) ListPackages() ([]string, error) {
	return h.PackageManager.ListPackages(h.UnixHost)
}

func (h MacOSHost) AddPackage(pkg string) error {
	return h.PackageManager.AddPackage(h.UnixHost, pkg)
}

func (h MacOSHost) RemovePackage(pkg string) error {
	return h.PackageManager.RemovePackage(h.UnixHost, pkg)
}

func (h MacOSHost) UpgradePackage(pkg string) error {
	return h.PackageManager.UpgradePackage(h.UnixHost, pkg)
}

func (h MacOSHost) CheckUpdates() ([]Update, error) {
	// Implement the update check for macOS hosts.
	return []Update{}, nil
}

func (h MacOSHost) RunCommand(cmd string) (string, error) {
	return h.UnixHost.RunCommand(cmd)
}

type HostOption func(*UnixHost)

func WithUser(user string) HostOption {
	return func(host *UnixHost) {
		host.User = user
	}
}

func WithPassword(password string) HostOption {
	return func(host *UnixHost) {
		host.Password = password
	}
}

func WithKeyPassphrase(keyPassphrase string) HostOption {
	return func(host *UnixHost) {
		host.KeyPassphrase = keyPassphrase
	}
}

func WithOS(os string) HostOption {
	return func(host *UnixHost) {
		host.OS = os
	}
}

func NewHost(hostname string, options ...HostOption) (Host, error) {
	host := &UnixHost{
		Hostname: hostname,
	}

	for _, option := range options {
		option(host)
	}

	// If the username has not been specified, use the current user's username.
	if host.User == "" {
		currentUser, err := user.Current()
		if err != nil {
			return nil, fmt.Errorf("could not get current user: %v", err)
		}
		host.User = currentUser.Username
	}

	// If the OS has not been specified, determine it.
	if host.OS == "" {
		os, err := determineOS(host)
		if err != nil {
			return nil, err
		}
		host.OS = os
	}

	switch host.OS {
	case "Linux":
		// Determine the package manager.
		// Here we just guess based on the contents of /etc/os-release.
		osRelease, _ := host.RunCommand("cat /etc/os-release")
		if strings.Contains(osRelease, "ID=ubuntu") || strings.Contains(osRelease, "ID=debian") {
			return LinuxHost{*host, AptPackageManager{}}, nil
		} else {
			// Assume Red Hat/CentOS/Fedora if not Debian/Ubuntu.
			return LinuxHost{*host, YumPackageManager{}}, nil
		}
	case "Darwin":
		return MacOSHost{*host, BrewPackageManager{}}, nil
	default:
		return nil, fmt.Errorf("unsupported operating system: %s", host.OS)
	}

}

func determineOS(host *UnixHost) (string, error) {
	output, err := host.RunCommand("uname")
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(output), nil
}

func (h UnixHost) RunCommand(cmd string) (string, error) {
	log.Printf("Running command '%s' on host '%s' with user '%s'\n", cmd, h.Hostname, h.User)
	// If the hostname is "localhost" or "127.0.0.1", run the command locally.
	if h.Hostname == "localhost" || h.Hostname == "127.0.0.1" {
		parts := strings.Fields(cmd)
		head := parts[0]
		parts = parts[1:]

		out, err := exec.Command(head, parts...).Output()
		if err != nil {
			return "", err
		}

		return string(out), nil
	}

	// Otherwise, run the command over SSH.
	var authMethod ssh.AuthMethod

	if h.Password != "" {
		log.Println("Using password authentication")
		authMethod = ssh.Password(h.Password)
	} else {
		log.Println("Using public key authentication")
		var keyManager SSHKeyManager
		if h.KeyPassphrase != "" {
			keyManager = FileSSHKeyManager{}
		} else {
			keyManager = AgentSSHKeyManager{}
		}

		keys, err := keyManager.ReadPrivateKeys(h.KeyPassphrase)
		if err != nil {
			return "", err
		}

		authMethod = ssh.PublicKeysCallback(func() ([]ssh.Signer, error) {
			return keys, nil
		})
	}

	config := &ssh.ClientConfig{
		User: h.User,
		Auth: []ssh.AuthMethod{
			authMethod,
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	client, err := ssh.Dial("tcp", h.Hostname+":22", config)
	if err != nil {
		return "", err
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	output, err := session.CombinedOutput(cmd)
	if err != nil {
		return "", err
	}

	return string(output), nil
}
