package steelcut

import (
	"fmt"
	"log"
	"strings"
)

// PackageManager interface defines the methods that package manager implementations must provide.
type PackageManager interface {
	ListPackages(*UnixHost) ([]string, error)
	AddPackage(*UnixHost, string) error
	RemovePackage(*UnixHost, string) error
	UpgradePackage(*UnixHost, string) error
	CheckOSUpdates(host *UnixHost) ([]string, error)
	UpgradeAll(*UnixHost) ([]Update, error)
}

// Update represents a package update.
type Update struct {
	PackageName string
	Version     string
}

type YumPackageManager struct {
	Executor CommandExecutor
	Logger   *log.Logger
}

func (pm YumPackageManager) ListPackages(host *UnixHost) ([]string, error) {
	output, err := pm.Executor.RunCommand("yum list installed", CommandOptions{UseSudo: false})
	if err != nil {
		return nil, err
	}
	return strings.Split(output, "\n"), nil
}

func (pm YumPackageManager) AddPackage(host *UnixHost, pkg string) error {
	_, err := pm.Executor.RunCommand(fmt.Sprintf("yum install -y %s", pkg), CommandOptions{UseSudo: true})
	return err
}

func (pm YumPackageManager) RemovePackage(host *UnixHost, pkg string) error {
	_, err := pm.Executor.RunCommand(fmt.Sprintf("yum remove -y %s", pkg), CommandOptions{UseSudo: true})
	return err
}

func (pm YumPackageManager) UpgradePackage(host *UnixHost, pkg string) error {
	_, err := pm.Executor.RunCommand(fmt.Sprintf("yum upgrade -y %s", pkg), CommandOptions{UseSudo: true})
	return err
}

func (pm YumPackageManager) CheckOSUpdates(host *UnixHost) ([]string, error) {
	log.Print("Checking for YUM OS updates")
	output, err := pm.Executor.RunCommand("yum check-update", CommandOptions{UseSudo: true})
	if err != nil {
		log.Printf("Error checking YUM updates: %v", err)
		return nil, err
	}

	updates := strings.Split(output, "\n")
	log.Printf("YUM Updates available: %v", updates)
	return updates, nil
}

// UpgradeAll upgrades all the packages to their latest versions.
func (pm YumPackageManager) UpgradeAll(host *UnixHost) ([]Update, error) {
	output, err := pm.Executor.RunCommand("yum update -y", CommandOptions{UseSudo: true})
	if err != nil {
		return nil, fmt.Errorf("failed to upgrade all packages: %v, Output: %s", err, output)
	}
	updates := parseUpdates(output)
	return updates, nil
}

type AptPackageManager struct {
	Executor CommandExecutor
	Logger   *log.Logger
}

// ListPackages returns the installed packages.
func (pm AptPackageManager) ListPackages(host *UnixHost) ([]string, error) {
	output, err := pm.Executor.RunCommand("apt list --installed", CommandOptions{UseSudo: false})
	if err != nil {
		return nil, err
	}

	packages := strings.Split(output, "\n")
	return packages, nil
}

// AddPackage adds a package to the host.
func (pm AptPackageManager) AddPackage(host *UnixHost, pkg string) error {
	_, err := pm.Executor.RunCommand(fmt.Sprintf("apt install -y %s", pkg), CommandOptions{UseSudo: true})
	return err
}

// RemovePackage removes a package from the host.
func (pm AptPackageManager) RemovePackage(host *UnixHost, pkg string) error {
	_, err := pm.Executor.RunCommand(fmt.Sprintf("apt remove -y %s", pkg), CommandOptions{UseSudo: true})
	return err
}

// UpgradePackage upgrades a package to the latest version.
func (pm AptPackageManager) UpgradePackage(host *UnixHost, pkg string) error {
	_, err := pm.Executor.RunCommand(fmt.Sprintf("apt upgrade -y %s", pkg), CommandOptions{UseSudo: true})
	return err
}

// UpgradeAll upgrades all the packages to their latest versions.
func (pm AptPackageManager) UpgradeAll(host *UnixHost) ([]Update, error) {
	output, err := pm.Executor.RunCommand("apt upgrade -y", CommandOptions{UseSudo: true})
	if err != nil {
		return nil, fmt.Errorf("failed to upgrade all packages: %v, Output: %s", err, output)
	}
	updates := pm.parseAptUpdates(output)
	return updates, nil
}

// CheckOSUpdates checks for OS updates.
func (pm AptPackageManager) CheckOSUpdates(host *UnixHost) ([]string, error) {
	_, err := pm.Executor.RunCommand("apt update", CommandOptions{UseSudo: true})
	if err != nil {
		log.Fatalf("Failed to update apt: %v", err)
		return nil, fmt.Errorf("Failed to update apt: %w", err)
	}

	output, err := pm.Executor.RunCommand("apt list --upgradable", CommandOptions{UseSudo: false})
	if err != nil {
		return nil, err
	}

	updates := strings.Split(output, "\n")
	return updates, nil
}

// parseAptUpdates parses the output of `apt upgrade -y` to get the list of packages that will be upgraded.
func (pm AptPackageManager) parseAptUpdates(output string) []Update {
	lines := strings.Split(output, "\n")
	var updates []Update

	for _, line := range lines {
		// Example line: "packagename/xenial 2.0.1 amd64 [upgradable from: 1.9.3]"
		parts := strings.Fields(line)
		if len(parts) < 5 || parts[4] != "[upgradable" {
			continue
		}

		packageName := strings.Split(parts[0], "/")[0]
		version := parts[1]
		update := Update{
			PackageName: packageName,
			Version:     version,
		}

		updates = append(updates, update)
	}

	return updates
}

type BrewPackageManager struct {
	Executor CommandExecutor
	Logger   *log.Logger
}

// ListPackages returns the installed packages.
func (pm BrewPackageManager) ListPackages(host *UnixHost) ([]string, error) {
	output, err := pm.Executor.RunCommand("brew list --version", CommandOptions{UseSudo: false})
	if err != nil {
		return nil, err
	}

	packages := strings.Split(output, "\n")
	return packages, nil
}

// AddPackage adds a package to the host.
func (pm BrewPackageManager) AddPackage(host *UnixHost, pkg string) error {
	_, err := pm.Executor.RunCommand(fmt.Sprintf("brew install %s", pkg), CommandOptions{UseSudo: false})
	return err
}

func (pm BrewPackageManager) RemovePackage(host *UnixHost, pkg string) error {
	_, err := pm.Executor.RunCommand(fmt.Sprintf("brew uninstall %s", pkg), CommandOptions{UseSudo: false})
	return err
}

// UpgradePackage upgrades a package to the latest version.
func (pm BrewPackageManager) UpgradePackage(host *UnixHost, pkg string) error {
	_, err := pm.Executor.RunCommand(fmt.Sprintf("brew upgrade %s", pkg), CommandOptions{UseSudo: false})
	return err
}

// CheckOSUpdates checks for OS updates.
func (pm BrewPackageManager) CheckOSUpdates(host *UnixHost) ([]string, error) {
	output, err := pm.Executor.RunCommand("brew outdated", CommandOptions{UseSudo: false})
	if err != nil {
		return nil, err
	}

	updates := strings.Split(output, "\n")
	return updates, nil
}

// UpgradeAll upgrades all the packages to their latest versions.
func (pm BrewPackageManager) UpgradeAll(host *UnixHost) ([]Update, error) {
	// We explicitly don't want to run as root here, as brew will complain
	output, err := pm.Executor.RunCommand("brew upgrade", CommandOptions{UseSudo: false})
	if err != nil {
		return nil, fmt.Errorf("failed to upgrade all packages: %v, Output: %s", err, output)
	}
	updates := pm.parseUpdates(output)
	return updates, nil
}

// parseUpdates parses the output of `brew upgrade` to get the list of packages that will be upgraded.
func (pm BrewPackageManager) parseUpdates(output string) []Update {
	lines := strings.Split(output, "\n")
	var updates []Update

	for _, line := range lines {
		parts := strings.Split(line, " ")
		if len(parts) < 2 {
			continue
		}

		update := Update{
			PackageName: parts[0],
			Version:     parts[1],
		}

		updates = append(updates, update)
	}

	return updates
}

// parseUpdates is a common function used to parse package update information.
func parseUpdates(output string) []Update {
	lines := strings.Split(output, "\n")
	var updates []Update

	for _, line := range lines {
		parts := strings.Split(line, " ")
		if len(parts) < 2 {
			continue
		}

		update := Update{
			PackageName: parts[0],
			Version:     parts[1],
		}

		updates = append(updates, update)
	}

	return updates
}
