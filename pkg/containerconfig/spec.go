package containerconfig

// ContainerSpec represents the configuration of a Docker container
type ContainerSpec struct {
	Name       string
	Image      string
	Env        []string
	Volumes    []string
	Ports      []string
	Networks   []string
	Command    []string
	WorkingDir string
	Labels     map[string]string
	EntryPoint []string
	Devices    []string
	ExtraHosts []string
	Restart    string
}

// RunOptions contains options for generating docker run command
type RunOptions struct {
	Name string
}
