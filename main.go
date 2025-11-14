package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/lhc03/docker-config-extractor/pkg/containerconfig"
)

// Manager handles Docker container operations with clean, single-responsibility methods
type Manager struct {
	containerName string
	devSwapDir    string
	logger        *log.Logger
}

// NewManager creates a new Manager instance with a logger
func NewManager(containerName, devSwapDir string) *Manager {
	return &Manager{
		containerName: containerName,
		devSwapDir:    devSwapDir,
		logger:        log.New(os.Stdout, "[Manager] ", log.LstdFlags),
	}
}

// CheckDevContainerExists checks if the dev container exists
func (m *Manager) CheckDevContainerExists(devContainerName string) (bool, error) {
	m.logger.Printf("Checking if dev container '%s' exists...", devContainerName)
	
	cmd := exec.Command("docker", "ps", "-a", "--filter", fmt.Sprintf("name=^%s$", devContainerName), "--format", "{{.Names}}")
	var out bytes.Buffer
	cmd.Stdout = &out
	
	if err := cmd.Run(); err != nil {
		return false, fmt.Errorf("failed to check container '%s': %w", devContainerName, err)
	}

	exists := strings.TrimSpace(out.String()) == devContainerName
	m.logger.Printf("Container '%s' exists: %v", devContainerName, exists)
	return exists, nil
}

// GetContainerConfig retrieves the container configuration using docker inspect
func (m *Manager) GetContainerConfig() (*containerconfig.ContainerSpec, error) {
	m.logger.Printf("Inspecting container '%s'...", m.containerName)
	
	cmd := exec.Command("docker", "inspect", m.containerName)
	var out bytes.Buffer
	var errOut bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errOut
	
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to inspect container '%s': %w, stderr: %s", m.containerName, err, errOut.String())
	}

	spec, err := containerconfig.ParseInspectJSON(out.String())
	if err != nil {
		return nil, fmt.Errorf("failed to parse inspect JSON for container '%s': %w", m.containerName, err)
	}

	m.logger.Printf("Successfully parsed container config for '%s'", m.containerName)
	return spec, nil
}

// CreateDevContainer creates a development container with additional dev tools
// This method separates docker run from docker exec operations
func (m *Manager) CreateDevContainer(devContainerName string, enableDebugger bool, injectScript string) error {
	m.logger.Printf("Starting creation of dev container '%s'...", devContainerName)
	
	// Step 1: Get original container config
	spec, err := m.GetContainerConfig()
	if err != nil {
		return fmt.Errorf("failed to get container config: %w", err)
	}

	// Step 2: Modify spec for dev container
	if m.devSwapDir != "" {
		m.logger.Printf("Adding dev-swap volume: %s:/dev-swap", m.devSwapDir)
		spec.Volumes = append(spec.Volumes, fmt.Sprintf("%s:/dev-swap", m.devSwapDir))
	}

	if enableDebugger {
		m.logger.Println("Adding debugger port: 2345:2345")
		spec.Ports = append(spec.Ports, "2345:2345")
	}

	// Step 3: Generate and execute docker run command
	opts := &containerconfig.RunOptions{
		Name: devContainerName,
	}
	runArgs := containerconfig.GenerateRunCommand(spec, opts)
	
	m.logger.Printf("Executing docker run command...")
	if err := m.executeDockerRun(runArgs); err != nil {
		return fmt.Errorf("failed to run dev container: %w", err)
	}

	// Step 4: Wait for container to be ready
	if err := m.waitForContainer(devContainerName, 10*time.Second); err != nil {
		return fmt.Errorf("container failed to start: %w", err)
	}

	// Step 5: Install debugger if requested
	if enableDebugger {
		if err := m.installDebugger(devContainerName); err != nil {
			m.logger.Printf("Warning: failed to install debugger: %v", err)
			// Don't fail the entire operation if debugger installation fails
		}
	}

	// Step 6: Inject custom script if provided
	if injectScript != "" {
		if err := m.executeInContainer(devContainerName, injectScript); err != nil {
			m.logger.Printf("Warning: failed to execute inject script: %v", err)
		}
	}

	m.logger.Printf("Dev container '%s' created successfully!", devContainerName)
	return nil
}

// executeDockerRun executes a docker run command (separated from docker exec)
func (m *Manager) executeDockerRun(args []string) error {
	m.logger.Println("Running docker run command...")
	
	cmd := exec.Command("docker", append([]string{"run", "-d"}, args...)...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker run failed: %w, stderr: %s", err, stderr.String())
	}
	
	m.logger.Printf("Container started: %s", strings.TrimSpace(stdout.String()))
	return nil
}

// waitForContainer waits for the container to be in running state
func (m *Manager) waitForContainer(containerName string, timeout time.Duration) error {
	m.logger.Printf("Waiting for container '%s' to be ready...", containerName)
	
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		cmd := exec.Command("docker", "inspect", "-f", "{{.State.Running}}", containerName)
		var out bytes.Buffer
		cmd.Stdout = &out
		
		if err := cmd.Run(); err == nil {
			if strings.TrimSpace(out.String()) == "true" {
				m.logger.Printf("Container '%s' is running", containerName)
				return nil
			}
		}
		
		time.Sleep(500 * time.Millisecond)
	}
	
	return fmt.Errorf("timeout waiting for container '%s' to start", containerName)
}

// installDebugger installs delve debugger in the container
func (m *Manager) installDebugger(containerName string) error {
	m.logger.Printf("Installing debugger in container '%s'...", containerName)
	
	// Step 1: Check if Go is installed
	checkGoCmd := exec.Command("docker", "exec", containerName, "which", "go")
	var checkOut bytes.Buffer
	checkGoCmd.Stdout = &checkOut
	
	if err := checkGoCmd.Run(); err != nil {
		return fmt.Errorf("Go is not installed in container '%s', cannot install debugger", containerName)
	}
	
	m.logger.Printf("Go found in container, proceeding with delve installation...")
	
	// Step 2: Install delve
	installCmd := exec.Command("docker", "exec", containerName, "go", "install", "github.com/go-delve/delve/cmd/dlv@latest")
	installCmd.Stdout = os.Stdout
	installCmd.Stderr = os.Stderr
	
	if err := installCmd.Run(); err != nil {
		return fmt.Errorf("failed to install delve: %w", err)
	}
	
	// Step 3: Verify delve installation
	verifyCmd := exec.Command("docker", "exec", containerName, "sh", "-c", "command -v dlv || echo 'dlv not found'")
	var verifyOut bytes.Buffer
	verifyCmd.Stdout = &verifyOut
	
	if err := verifyCmd.Run(); err != nil {
		return fmt.Errorf("failed to verify delve installation: %w", err)
	}
	
	if strings.Contains(verifyOut.String(), "not found") {
		return fmt.Errorf("delve installed but not found in PATH")
	}
	
	m.logger.Printf("Delve debugger installed successfully in '%s'", containerName)
	return nil
}

// executeInContainer executes a command inside the container using docker exec
func (m *Manager) executeInContainer(containerName, command string) error {
	m.logger.Printf("Executing command in container '%s': %s", containerName, command)
	
	cmd := exec.Command("docker", "exec", containerName, "sh", "-c", command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to execute command in container '%s': %w", containerName, err)
	}
	
	return nil
}

// StopDevContainer stops the dev container
func (m *Manager) StopDevContainer(devContainerName string) error {
	m.logger.Printf("Stopping container '%s'...", devContainerName)
	
	cmd := exec.Command("docker", "stop", devContainerName)
	var errOut bytes.Buffer
	cmd.Stderr = &errOut
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to stop container '%s': %w, stderr: %s", devContainerName, err, errOut.String())
	}
	
	m.logger.Printf("Container '%s' stopped successfully", devContainerName)
	return nil
}

// RemoveDevContainer removes the dev container
func (m *Manager) RemoveDevContainer(devContainerName string) error {
	m.logger.Printf("Removing container '%s'...", devContainerName)
	
	cmd := exec.Command("docker", "rm", devContainerName)
	var errOut bytes.Buffer
	cmd.Stderr = &errOut
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to remove container '%s': %w, stderr: %s", devContainerName, err, errOut.String())
	}
	
	m.logger.Printf("Container '%s' removed successfully", devContainerName)
	return nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: docker-config-extractor <container-name> [dev-container-name] [dev-swap-dir]")
		fmt.Println("\nExample:")
		fmt.Println("  docker-config-extractor myapp myapp-dev /path/to/dev-swap")
		os.Exit(1)
	}

	containerName := os.Args[1]
	devContainerName := containerName + "-dev"
	devSwapDir := ""

	if len(os.Args) >= 3 {
		devContainerName = os.Args[2]
	}
	if len(os.Args) >= 4 {
		devSwapDir = os.Args[3]
	}

	manager := NewManager(containerName, devSwapDir)

	// Check if dev container already exists
	exists, err := manager.CheckDevContainerExists(devContainerName)
	if err != nil {
		log.Fatalf("Error checking dev container: %v", err)
	}

	if exists {
		fmt.Printf("\nDev container '%s' already exists.\n", devContainerName)
		fmt.Print("Do you want to recreate it? (y/n): ")
		var answer string
		fmt.Scanln(&answer)
		
		if strings.ToLower(strings.TrimSpace(answer)) == "y" {
			if err := manager.StopDevContainer(devContainerName); err != nil {
				log.Printf("Warning: error stopping container: %v", err)
			}
			if err := manager.RemoveDevContainer(devContainerName); err != nil {
				log.Fatalf("Error removing container: %v", err)
			}
		} else {
			fmt.Println("Exiting without changes.")
			return
		}
	}

	// Create dev container with debugger support
	enableDebugger := true
	injectScript := "echo 'Dev container is ready for development!'"
	
	if err := manager.CreateDevContainer(devContainerName, enableDebugger, injectScript); err != nil {
		log.Fatalf("Error creating dev container: %v", err)
	}

	fmt.Printf("\n✓ Dev container '%s' is ready!\n", devContainerName)
	fmt.Println("\nYou can now:")
	fmt.Printf("  - Attach to it: docker exec -it %s /bin/sh\n", devContainerName)
	fmt.Printf("  - Debug with delve on port 2345\n")
}
