package containerconfig

import (
	"fmt"
)

// GenerateRunCommand generates docker run arguments from ContainerSpec
// Returns a slice of arguments (without "docker" and "run")
func GenerateRunCommand(spec *ContainerSpec, opts *RunOptions) []string {
	var args []string

	// Add name
	if opts != nil && opts.Name != "" {
		args = append(args, "--name", opts.Name)
	} else if spec.Name != "" {
		args = append(args, "--name", spec.Name)
	}

	// Add environment variables
	for _, env := range spec.Env {
		args = append(args, "-e", env)
	}

	// Add volumes
	for _, vol := range spec.Volumes {
		args = append(args, "-v", vol)
	}

	// Add ports
	for _, port := range spec.Ports {
		args = append(args, "-p", port)
	}

	// Add networks
	for _, network := range spec.Networks {
		args = append(args, "--network", network)
	}

	// Add working directory
	if spec.WorkingDir != "" {
		args = append(args, "-w", spec.WorkingDir)
	}

	// Add labels
	for key, value := range spec.Labels {
		args = append(args, "-l", fmt.Sprintf("%s=%s", key, value))
	}

	// Add devices
	for _, device := range spec.Devices {
		args = append(args, "--device", device)
	}

	// Add extra hosts
	for _, host := range spec.ExtraHosts {
		args = append(args, "--add-host", host)
	}

	// Add restart policy
	if spec.Restart != "" {
		args = append(args, "--restart", spec.Restart)
	}

	// Add entrypoint
	if len(spec.EntryPoint) > 0 {
		args = append(args, "--entrypoint", spec.EntryPoint[0])
	}

	// Add image
	args = append(args, spec.Image)

	// Add command arguments
	if len(spec.Command) > 0 {
		args = append(args, spec.Command...)
	}

	return args
}
