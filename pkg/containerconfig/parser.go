package containerconfig

import (
	"encoding/json"
	"fmt"
	"strings"
)

// InspectData represents the structure of docker inspect JSON output
type InspectData struct {
	Name   string `json:"Name"`
	Config struct {
		Image      string            `json:"Image"`
		Env        []string          `json:"Env"`
		Cmd        []string          `json:"Cmd"`
		Entrypoint []string          `json:"Entrypoint"`
		Labels     map[string]string `json:"Labels"`
		WorkingDir string            `json:"WorkingDir"`
	} `json:"Config"`
	Mounts []struct {
		Type        string `json:"Type"`
		Source      string `json:"Source"`
		Destination string `json:"Destination"`
		Mode        string `json:"Mode"`
		RW          bool   `json:"RW"`
	} `json:"Mounts"`
	NetworkSettings struct {
		Networks map[string]interface{} `json:"Networks"`
		Ports    map[string][]struct {
			HostIP   string `json:"HostIp"`
			HostPort string `json:"HostPort"`
		} `json:"Ports"`
	} `json:"NetworkSettings"`
	HostConfig struct {
		Devices []struct {
			PathOnHost        string `json:"PathOnHost"`
			PathInContainer   string `json:"PathInContainer"`
			CgroupPermissions string `json:"CgroupPermissions"`
		} `json:"Devices"`
		RestartPolicy struct {
			Name              string `json:"Name"`
			MaximumRetryCount int    `json:"MaximumRetryCount"`
		} `json:"RestartPolicy"`
		ExtraHosts []string `json:"ExtraHosts"`
	} `json:"HostConfig"`
}

// ParseInspectJSON parses docker inspect JSON output and returns ContainerSpec
func ParseInspectJSON(jsonData string) (*ContainerSpec, error) {
	var inspectArray []InspectData
	if err := json.Unmarshal([]byte(jsonData), &inspectArray); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	if len(inspectArray) == 0 {
		return nil, fmt.Errorf("empty inspect data")
	}

	data := inspectArray[0]
	spec := &ContainerSpec{
		Name:       strings.TrimPrefix(data.Name, "/"),
		Image:      data.Config.Image,
		Env:        data.Config.Env,
		Command:    data.Config.Cmd,
		EntryPoint: data.Config.Entrypoint,
		Labels:     data.Config.Labels,
		WorkingDir: data.Config.WorkingDir,
	}

	// Parse volumes from mounts
	for _, mount := range data.Mounts {
		var volumeStr string
		if mount.Type == "bind" {
			volumeStr = fmt.Sprintf("%s:%s", mount.Source, mount.Destination)
		} else if mount.Type == "volume" {
			volumeStr = fmt.Sprintf("%s:%s", mount.Source, mount.Destination)
		}
		if volumeStr != "" {
			if !mount.RW {
				volumeStr += ":ro"
			}
			spec.Volumes = append(spec.Volumes, volumeStr)
		}
	}

	// Parse ports
	for containerPort, bindings := range data.NetworkSettings.Ports {
		if len(bindings) > 0 {
			for _, binding := range bindings {
				if binding.HostPort != "" {
					portStr := fmt.Sprintf("%s:%s", binding.HostPort, strings.Split(containerPort, "/")[0])
					spec.Ports = append(spec.Ports, portStr)
				}
			}
		}
	}

	// Parse networks
	for networkName := range data.NetworkSettings.Networks {
		spec.Networks = append(spec.Networks, networkName)
	}

	// Parse devices
	for _, device := range data.HostConfig.Devices {
		deviceStr := fmt.Sprintf("%s:%s", device.PathOnHost, device.PathInContainer)
		spec.Devices = append(spec.Devices, deviceStr)
	}

	// Parse restart policy
	if data.HostConfig.RestartPolicy.Name != "" && data.HostConfig.RestartPolicy.Name != "no" {
		spec.Restart = data.HostConfig.RestartPolicy.Name
	}

	// Parse extra hosts
	spec.ExtraHosts = data.HostConfig.ExtraHosts

	return spec, nil
}
