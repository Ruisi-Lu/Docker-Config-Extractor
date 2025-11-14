package containerconfig

import (
	"fmt"
	"strings"
)

// GenerateRunCommand generates a docker run command from ContainerSpec
func GenerateRunCommand(spec *ContainerSpec, opts *RunOptions) string {
	var parts []string
	parts = append(parts, "docker run")

	// Add name
	if opts != nil && opts.Name != "" {
		parts = append(parts, fmt.Sprintf("--name %s", opts.Name))
	} else if spec.Name != "" {
		parts = append(parts, fmt.Sprintf("--name %s", spec.Name))
	}

	// Add environment variables
	for _, env := range spec.Env {
		parts = append(parts, fmt.Sprintf("-e %q", env))
	}

	// Add volumes
	for _, vol := range spec.Volumes {
		parts = append(parts, fmt.Sprintf("-v %s", vol))
	}

	// Add ports
	for _, port := range spec.Ports {
		parts = append(parts, fmt.Sprintf("-p %s", port))
	}

	// Add networks
	for _, network := range spec.Networks {
		parts = append(parts, fmt.Sprintf("--network %s", network))
	}

	// Add working directory
	if spec.WorkingDir != "" {
		parts = append(parts, fmt.Sprintf("-w %s", spec.WorkingDir))
	}

	// Add labels
	for key, value := range spec.Labels {
		parts = append(parts, fmt.Sprintf("-l %s=%q", key, value))
	}

	// Add devices
	for _, device := range spec.Devices {
		parts = append(parts, fmt.Sprintf("--device %s", device))
	}

	// Add extra hosts
	for _, host := range spec.ExtraHosts {
		parts = append(parts, fmt.Sprintf("--add-host %s", host))
	}

	// Add restart policy
	if spec.Restart != "" {
		parts = append(parts, fmt.Sprintf("--restart %s", spec.Restart))
	}

	// Add entrypoint
	if len(spec.EntryPoint) > 0 {
		// Use only the first element as entrypoint executable
		parts = append(parts, fmt.Sprintf("--entrypoint %s", spec.EntryPoint[0]))
	}

	// Add image
	parts = append(parts, spec.Image)

	// Add command
	if len(spec.Command) > 0 {
		parts = append(parts, strings.Join(spec.Command, " "))
	}

	return strings.Join(parts, " ")
}
