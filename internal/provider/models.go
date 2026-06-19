// Copyright 536 Technologies LLC
// SPDX-License-Identifier: MPL-2.0

package provider

import "encoding/json"

type sandboxCreateRequest struct {
	TemplateID           string                `json:"templateID"`
	Timeout              *int64                `json:"timeout,omitempty"`
	AutoPause            *bool                 `json:"autoPause,omitempty"`
	AutoResume           *sandboxAutoResume    `json:"autoResume,omitempty"`
	Secure               *bool                 `json:"secure,omitempty"`
	AllowInternetAccess  *bool                 `json:"allow_internet_access,omitempty"`
	Network              *sandboxNetworkConfig `json:"network,omitempty"`
	Metadata             map[string]string     `json:"metadata,omitempty"`
	EnvironmentVariables map[string]string     `json:"envVars,omitempty"`
	VolumeMounts         []sandboxVolumeMount  `json:"volumeMounts,omitempty"`
}

type sandboxAutoResume struct {
	Enabled bool `json:"enabled"`
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
	Lifecycle           *sandboxLifecycle     `json:"lifecycle"`
	VolumeMounts        []sandboxVolumeMount  `json:"volumeMounts"`
}

type listedSandboxResponse struct {
	TemplateID   string               `json:"templateID"`
	SandboxID    string               `json:"sandboxID"`
	Alias        string               `json:"alias"`
	ClientID     string               `json:"clientID"`
	StartedAt    string               `json:"startedAt"`
	EndAt        string               `json:"endAt"`
	CPUCount     *int64               `json:"cpuCount"`
	MemoryMB     *int64               `json:"memoryMB"`
	DiskSizeMB   *int64               `json:"diskSizeMB"`
	Metadata     map[string]string    `json:"metadata"`
	State        string               `json:"state"`
	EnvdVersion  string               `json:"envdVersion"`
	VolumeMounts []sandboxVolumeMount `json:"volumeMounts"`
}

type sandboxNetworkConfig struct {
	AllowPublicTraffic *bool    `json:"allowPublicTraffic,omitempty"`
	AllowOut           []string `json:"allowOut,omitempty"`
	DenyOut            []string `json:"denyOut,omitempty"`
	MaskRequestHost    string   `json:"maskRequestHost,omitempty"`
}

type sandboxLifecycle struct {
	AutoResume bool   `json:"autoResume"`
	OnTimeout  string `json:"onTimeout"`
}

type sandboxVolumeMount struct {
	Name string `json:"name"`
	Path string `json:"path"`
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

type teamResponse struct {
	TeamID    string `json:"teamID"`
	Name      string `json:"name"`
	APIKey    string `json:"apiKey"`
	IsDefault bool   `json:"isDefault"`
}

type teamMetricResponse struct {
	Timestamp           string  `json:"timestamp"`
	TimestampUnix       int64   `json:"timestampUnix"`
	ConcurrentSandboxes int64   `json:"concurrentSandboxes"`
	SandboxStartRate    float64 `json:"sandboxStartRate"`
}

type maxTeamMetricResponse struct {
	Timestamp     string  `json:"timestamp"`
	TimestampUnix int64   `json:"timestampUnix"`
	Value         float64 `json:"value"`
}

type maskingDetailsResponse struct {
	Prefix            string `json:"prefix"`
	ValueLength       int64  `json:"valueLength"`
	MaskedValuePrefix string `json:"maskedValuePrefix"`
	MaskedValueSuffix string `json:"maskedValueSuffix"`
}

type teamUserResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

type apiKeyRequest struct {
	Name string `json:"name"`
}

type apiKeyResponse struct {
	ID        string                  `json:"id"`
	Name      string                  `json:"name"`
	Key       string                  `json:"key"`
	Mask      *maskingDetailsResponse `json:"mask"`
	CreatedAt string                  `json:"createdAt"`
	CreatedBy *teamUserResponse       `json:"createdBy"`
	LastUsed  *string                 `json:"lastUsed"`
}

type accessTokenRequest struct {
	Name string `json:"name"`
}

type accessTokenResponse struct {
	ID        string                  `json:"id"`
	Name      string                  `json:"name"`
	Token     string                  `json:"token"`
	Mask      *maskingDetailsResponse `json:"mask"`
	CreatedAt string                  `json:"createdAt"`
}

type snapshotRequest struct {
	Name string `json:"name,omitempty"`
}

type templateTagsAssignRequest struct {
	Target string   `json:"target"`
	Tags   []string `json:"tags"`
}

type templateTagsDeleteRequest struct {
	Name string   `json:"name"`
	Tags []string `json:"tags"`
}

type templateTagsAssignResponse struct {
	Tags    []string `json:"tags"`
	BuildID string   `json:"buildID"`
}

type templateTagResponse struct {
	Tag       string `json:"tag"`
	BuildID   string `json:"buildID"`
	CreatedAt string `json:"createdAt"`
}

type lifecycleWebhookRequest struct {
	Name            string   `json:"name,omitempty"`
	URL             string   `json:"url,omitempty"`
	Enabled         *bool    `json:"enabled,omitempty"`
	Events          []string `json:"events,omitempty"`
	SignatureSecret string   `json:"signatureSecret,omitempty"`
}

type lifecycleWebhookResponse struct {
	ID        string   `json:"id"`
	TeamID    string   `json:"teamId"`
	Name      string   `json:"name"`
	CreatedAt string   `json:"createdAt"`
	Enabled   bool     `json:"enabled"`
	URL       string   `json:"url"`
	Events    []string `json:"events"`
}

type lifecycleEventResponse struct {
	Version            string          `json:"version"`
	ID                 string          `json:"id"`
	Type               string          `json:"type"`
	EventData          json.RawMessage `json:"eventData"`
	SandboxBuildID     string          `json:"sandboxBuildId"`
	SandboxExecutionID string          `json:"sandboxExecutionId"`
	SandboxID          string          `json:"sandboxId"`
	SandboxTeamID      string          `json:"sandboxTeamId"`
	SandboxTemplateID  string          `json:"sandboxTemplateId"`
	Timestamp          string          `json:"timestamp"`
}

type snapshotResponse struct {
	SnapshotID string   `json:"snapshotID"`
	Names      []string `json:"names"`
}
