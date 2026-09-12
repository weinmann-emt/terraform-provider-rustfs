package rustfs

import (
	"context"
	"encoding/json"
)

// ModuleSwitchSource indicates where the effective module switch state came from.
// Values mirror the server's serialization: "env", "console", or "default".
type ModuleSwitchSource string

const (
	ModuleSwitchSourceEnv     ModuleSwitchSource = "env"
	ModuleSwitchSourceConsole ModuleSwitchSource = "console"
	ModuleSwitchSourceDefault ModuleSwitchSource = "default"
)

// ModuleSwitchState is the full server representation of the feature module
// switches. NotifyEnabled and AuditEnabled are the writable toggles; the
// remaining fields are read-only and describe the persisted state, the source
// of the effective value, and the admin discovery routes.
type ModuleSwitchState struct {
	NotifyEnabled bool `json:"notify_enabled"`
	AuditEnabled  bool `json:"audit_enabled"`
	// Read-only fields reported by GET.
	PersistedNotifyEnabled bool                   `json:"persisted_notify_enabled,omitempty"`
	PersistedAuditEnabled  bool                   `json:"persisted_audit_enabled,omitempty"`
	NotifySource           ModuleSwitchSource     `json:"notify_source,omitempty"`
	AuditSource            ModuleSwitchSource     `json:"audit_source,omitempty"`
	AdminDiscovery         *ModuleSwitchDiscovery `json:"admin_discovery,omitempty"`
}

// ModuleSwitchDiscovery holds the admin discovery routes exposed by the server.
type ModuleSwitchDiscovery struct {
	RuntimeCapabilities string `json:"runtimeCapabilities"`
	ClusterSnapshot     string `json:"clusterSnapshot"`
	ExtensionsCatalog   string `json:"extensionsCatalog"`
}

// ModuleSwitchUpdate is the request body for PUT /module-switches.
type ModuleSwitchUpdate struct {
	NotifyEnabled bool `json:"notify_enabled"`
	AuditEnabled  bool `json:"audit_enabled"`
}

// GetModuleSwitches reads the current feature module switches from the server.
func (c *RustfsAdmin) GetModuleSwitches() (*ModuleSwitchState, error) {
	reqData := RequestData{
		Method:  "GET",
		RelPath: "module-switches",
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	resp, err := c.doRequest(ctx, reqData)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	state := &ModuleSwitchState{}
	err = json.NewDecoder(resp.Body).Decode(state)
	return state, err
}

// SetModuleSwitches applies the feature module switch toggles. The server
// returns the resulting state after applying the update.
func (c *RustfsAdmin) SetModuleSwitches(update ModuleSwitchUpdate) (*ModuleSwitchState, error) {
	bytes, err := json.Marshal(update)
	if err != nil {
		return nil, err
	}
	reqData := RequestData{
		Method:  "PUT",
		RelPath: "module-switches",
		Content: bytes,
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	resp, err := c.doRequest(ctx, reqData)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	state := &ModuleSwitchState{}
	err = json.NewDecoder(resp.Body).Decode(state)
	return state, err
}
