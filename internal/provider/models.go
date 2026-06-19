// Copyright 536 Technologies LLC
// SPDX-License-Identifier: MPL-2.0

package provider

type sandboxCreateRequest struct {
	TemplateID           string                `json:"templateID"`
	Timeout              *int64                `json:"timeout,omitempty"`
	AutoPause            *bool                 `json:"autoPause,omitempty"`
	Secure               *bool                 `json:"secure,omitempty"`
	AllowInternetAccess  *bool                 `json:"allow_internet_access,omitempty"`
	Network              *sandboxNetworkConfig `json:"network,omitempty"`
	Metadata             map[string]string     `json:"metadata,omitempty"`
	EnvironmentVariables map[string]string     `json:"envVars,omitempty"`
}

type sandboxResponse struct {
	TemplateID         string  `json:"templateID"`
	SandboxID          string  `json:"sandboxID"`
	Alias              string  `json:"alias"`
	ClientID           string  `json:"clientID"`
	EnvdVersion        string  `json:"envdVersion"`
	EnvdAccessToken    *string `json:"envdAccessToken"`
	TrafficAccessToken *string `json:"trafficAccessToken"`
}

type sandboxDetailResponse struct {
	TemplateID          string                `json:"templateID"`
	SandboxID           string                `json:"sandboxID"`
	Alias               string                `json:"alias"`
	ClientID            string                `json:"clientID"`
	StartedAt           string                `json:"startedAt"`
	EndAt               string                `json:"endAt"`
	EnvdVersion         string                `json:"envdVersion"`
	EnvdAccessToken     *string               `json:"envdAccessToken"`
	TrafficAccessToken  *string               `json:"trafficAccessToken"`
	AllowInternetAccess *bool                 `json:"allowInternetAccess"`
	CPUCount            *int64                `json:"cpuCount"`
	MemoryMB            *int64                `json:"memoryMB"`
	DiskSizeMB          *int64                `json:"diskSizeMB"`
	Metadata            map[string]string     `json:"metadata"`
	State               string                `json:"state"`
	Network             *sandboxNetworkConfig `json:"network"`
}

type listedSandboxResponse struct {
	TemplateID  string            `json:"templateID"`
	SandboxID   string            `json:"sandboxID"`
	Alias       string            `json:"alias"`
	ClientID    string            `json:"clientID"`
	StartedAt   string            `json:"startedAt"`
	EndAt       string            `json:"endAt"`
	CPUCount    *int64            `json:"cpuCount"`
	MemoryMB    *int64            `json:"memoryMB"`
	DiskSizeMB  *int64            `json:"diskSizeMB"`
	Metadata    map[string]string `json:"metadata"`
	State       string            `json:"state"`
	EnvdVersion string            `json:"envdVersion"`
}

type sandboxNetworkConfig struct {
	AllowOut []string `json:"allowOut,omitempty"`
	DenyOut  []string `json:"denyOut,omitempty"`
}

type volumeRequest struct {
	Name string `json:"name"`
}

type volumeResponse struct {
	VolumeID string `json:"volumeID"`
	Name     string `json:"name"`
}

type templateResponse struct {
	TemplateID    string   `json:"templateID"`
	BuildID       string   `json:"buildID"`
	CPUCount      int64    `json:"cpuCount"`
	MemoryMB      int64    `json:"memoryMB"`
	DiskSizeMB    int64    `json:"diskSizeMB"`
	Public        bool     `json:"public"`
	Aliases       []string `json:"aliases"`
	Names         []string `json:"names"`
	CreatedAt     string   `json:"createdAt"`
	UpdatedAt     string   `json:"updatedAt"`
	LastSpawnedAt *string  `json:"lastSpawnedAt"`
	SpawnCount    int64    `json:"spawnCount"`
	BuildCount    int64    `json:"buildCount"`
	EnvdVersion   string   `json:"envdVersion"`
	BuildStatus   string   `json:"buildStatus"`
}

type templateCreateRequest struct {
	Name     string   `json:"name"`
	CPUCount *int64   `json:"cpuCount,omitempty"`
	MemoryMB *int64   `json:"memoryMB,omitempty"`
	Tags     []string `json:"tags,omitempty"`
}

type templateCreateResponse struct {
	TemplateID string   `json:"templateID"`
	BuildID    string   `json:"buildID"`
	Public     bool     `json:"public"`
	Aliases    []string `json:"aliases"`
	Names      []string `json:"names"`
	Tags       []string `json:"tags"`
}

type templateBuildStartRequest struct {
	FromImage string `json:"fromImage,omitempty"`
	Force     *bool  `json:"force,omitempty"`
	StartCmd  string `json:"startCmd,omitempty"`
	ReadyCmd  string `json:"readyCmd,omitempty"`
}

type templateBuildStatusResponse struct {
	TemplateID string                     `json:"templateID"`
	BuildID    string                     `json:"buildID"`
	Status     string                     `json:"status"`
	Logs       []string                   `json:"logs"`
	Reason     *templateBuildStatusReason `json:"reason"`
	LogEntries []templateBuildLogEntry    `json:"logEntries"`
}

type templateBuildStatusReason struct {
	Message string                  `json:"message"`
	Step    string                  `json:"step"`
	Logs    []templateBuildLogEntry `json:"logEntries"`
}

type templateBuildLogEntry struct {
	Timestamp string `json:"timestamp"`
	Message   string `json:"message"`
	Level     string `json:"level"`
	Step      string `json:"step"`
}
