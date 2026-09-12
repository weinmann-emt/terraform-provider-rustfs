package rustfs

import (
	"context"
	"encoding/json"
	"errors"
)

// KmsKey describes a KMS master key as returned by the KMS admin API.
//
// The wire shape is the key_metadata object of the create/describe/lifecycle
// responses. `key_state` is the authoritative lifecycle state (Enabled,
// Disabled or PendingDeletion).
type KmsKey struct {
	KeyID        string            `json:"key_id"`
	KeyState     string            `json:"key_state"`
	KeyUsage     string            `json:"key_usage"`
	Description  *string           `json:"description"`
	CreationDate string            `json:"creation_date"`
	DeletionDate *string           `json:"deletion_date"`
	Origin       string            `json:"origin"`
	KeyManager   string            `json:"key_manager"`
	Tags         map[string]string `json:"tags"`
}

// KmsKeyInfo is a single entry in a key listing response.
type KmsKeyInfo struct {
	KeyID       string            `json:"key_id"`
	Description *string           `json:"description"`
	Algorithm   string            `json:"algorithm"`
	Usage       string            `json:"usage"`
	Status      string            `json:"status"`
	Tags        map[string]string `json:"tags"`
	CreatedAt   string            `json:"created_at"`
	CreatedBy   *string           `json:"created_by"`
}

type kmsCreateResponse struct {
	Success     bool    `json:"success"`
	Message     string  `json:"message"`
	KeyID       string  `json:"key_id"`
	KeyMetadata *KmsKey `json:"key_metadata"`
}

type kmsDescribeResponse struct {
	Success     bool    `json:"success"`
	Message     string  `json:"message"`
	KeyMetadata *KmsKey `json:"key_metadata"`
}

type kmsListResponse struct {
	Success    bool         `json:"success"`
	Message    string       `json:"message"`
	Keys       []KmsKeyInfo `json:"keys"`
	Truncated  bool         `json:"truncated"`
	NextMarker *string      `json:"next_marker"`
}

type kmsLifecycleResponse struct {
	Success     bool    `json:"success"`
	Message     string  `json:"message"`
	KeyID       string  `json:"key_id"`
	KeyMetadata *KmsKey `json:"key_metadata"`
}

// CreateKmsKey creates a new KMS master key with the given name.
//
// The key name is carried as the `name` tag, which the RustFS create handler
// turns into the key id for backends that support named keys (Local backend)
// or into a server-generated key id for production backends.
func (c *RustfsAdmin) CreateKmsKey(name string) (KmsKey, error) {
	body := struct {
		Tags map[string]string `json:"tags"`
	}{Tags: map[string]string{"name": name}}
	//#nosec G117 — a KMS key name is a public identifier, not a secret
	bytes, err := json.Marshal(body)
	if err != nil {
		return KmsKey{}, err
	}
	reqData := RequestData{
		Method:  "POST",
		RelPath: "kms/keys",
		Content: bytes,
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	resp, err := c.doRequest(ctx, reqData)
	if err != nil {
		return KmsKey{}, err
	}
	defer resp.Body.Close()
	var out kmsCreateResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return KmsKey{}, err
	}
	if out.KeyMetadata != nil {
		return *out.KeyMetadata, nil
	}
	return KmsKey{KeyID: out.KeyID, KeyState: "Enabled", Tags: body.Tags}, nil
}

// DescribeKmsKey describes a KMS master key by its key id.
func (c *RustfsAdmin) DescribeKmsKey(keyID string) (KmsKey, error) {
	reqData := RequestData{
		Method:  "GET",
		RelPath: "kms/keys/" + keyID,
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	resp, err := c.doRequest(ctx, reqData)
	if err != nil {
		return KmsKey{}, err
	}
	defer resp.Body.Close()
	var out kmsDescribeResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return KmsKey{}, err
	}
	if out.KeyMetadata == nil {
		return KmsKey{}, errors.New(out.Message)
	}
	return *out.KeyMetadata, nil
}

// ListKmsKeys lists the KMS master keys known to the server.
func (c *RustfsAdmin) ListKmsKeys() ([]KmsKeyInfo, error) {
	reqData := RequestData{
		Method:  "GET",
		RelPath: "kms/keys",
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	resp, err := c.doRequest(ctx, reqData)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var out kmsListResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return out.Keys, nil
}

// EnableKmsKey enables a KMS master key.
func (c *RustfsAdmin) EnableKmsKey(keyID string) error {
	return c.setKmsKeyState(keyID, "kms/keys/enable")
}

// DisableKmsKey disables a KMS master key.
func (c *RustfsAdmin) DisableKmsKey(keyID string) error {
	return c.setKmsKeyState(keyID, "kms/keys/disable")
}

func (c *RustfsAdmin) setKmsKeyState(keyID, relPath string) error {
	body := struct {
		KeyID string `json:"key_id"`
	}{KeyID: keyID}
	//#nosec G117 — a KMS key id is a public identifier, not a secret
	bytes, err := json.Marshal(body)
	if err != nil {
		return err
	}
	reqData := RequestData{
		Method:  "POST",
		RelPath: relPath,
		Content: bytes,
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	resp, err := c.doRequest(ctx, reqData)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var out kmsLifecycleResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return err
	}
	if !out.Success {
		return errors.New(out.Message)
	}
	return nil
}

// RotateKmsKey rotates a KMS master key and returns the key's new metadata.
//
// Note that the Local development backend does not support rotation: the
// server answers with a 501 (KmsError::UnsupportedCapability) for backends
// whose capabilities report rotate=false.
func (c *RustfsAdmin) RotateKmsKey(keyID string) (KmsKey, error) {
	body := struct {
		KeyID string `json:"key_id"`
	}{KeyID: keyID}
	//#nosec G117 — a KMS key id is a public identifier, not a secret
	bytes, err := json.Marshal(body)
	if err != nil {
		return KmsKey{}, err
	}
	reqData := RequestData{
		Method:  "POST",
		RelPath: "kms/keys/rotate",
		Content: bytes,
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	resp, err := c.doRequest(ctx, reqData)
	if err != nil {
		return KmsKey{}, err
	}
	defer resp.Body.Close()
	var out kmsLifecycleResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return KmsKey{}, err
	}
	if out.KeyMetadata == nil {
		return KmsKey{}, errors.New(out.Message)
	}
	return *out.KeyMetadata, nil
}

// DeleteKmsKey schedules the deletion of a KMS master key.
//
// The RustFS admin API marks this route Critical: destroying a master key
// makes every object encrypted under it permanently unreadable. The server
// schedules the deletion behind a pending window (default 30 days) unless
// force_immediate/confirm_key_id are supplied; this client deliberately never
// sends them, so callers always get the cancellable scheduled deletion.
func (c *RustfsAdmin) DeleteKmsKey(keyID string) error {
	body := struct {
		KeyID string `json:"key_id"`
	}{KeyID: keyID}
	//#nosec G117 — a KMS key id is a public identifier, not a secret
	bytes, err := json.Marshal(body)
	if err != nil {
		return err
	}
	reqData := RequestData{
		Method:  "DELETE",
		RelPath: "kms/keys/delete",
		Content: bytes,
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	resp, err := c.doRequest(ctx, reqData)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}
