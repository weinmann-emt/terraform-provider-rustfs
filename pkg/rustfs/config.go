package rustfs

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strings"
)

// ConfigKV is a single key/value pair of a sub-system config scope.
type ConfigKV struct {
	Key   string
	Value string
}

// configHelpResponse is the JSON envelope returned by help-config-kv.
type configHelpResponse struct {
	SubSys          string `json:"subSys"`
	Description     string `json:"description"`
	MultipleTargets bool   `json:"multipleTargets"`
	KeysHelp        []struct {
		Key             string `json:"key"`
		Type            string `json:"type"`
		Description     string `json:"description"`
		Optional        bool   `json:"optional"`
		MultipleTargets bool   `json:"multipleTargets"`
	} `json:"keysHelp"`
}

// GetConfig returns the current KV settings of a config scope from GET get-config-kv.
//
// subSystem is a config scope: either a bare sub-system (e.g. "scanner") or a
// targeted sub-system (e.g. "notify_webhook:primary"). The server renders the
// scope as text lines in the form `scope key="value" key2="value2"`.
func (c *RustfsAdmin) GetConfig(subSystem string) ([]ConfigKV, error) {
	query := make(url.Values)
	query.Set("key", subSystem)
	reqData := RequestData{
		Method:      "GET",
		RelPath:     "get-config-kv",
		QueryValues: query,
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	resp, err := c.doRequest(ctx, reqData)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return parseConfigKVs(string(body))
}

// SetConfig sets the KV settings of a config scope via PUT set-config-kv.
//
// The request body is a plain text config directive in the same format the
// server renders: `scope key="value" key2="value2"`.
func (c *RustfsAdmin) SetConfig(subSystem string, kvs []ConfigKV) error {
	reqData := RequestData{
		Method:  "PUT",
		RelPath: "set-config-kv",
		Content: []byte(buildConfigDirective(subSystem, kvs)),
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_, err := c.doRequest(ctx, reqData)
	return err
}

// DeleteConfig removes a config scope via DELETE del-config-kv. The whole
// sub-system or the named target is removed; keys that are not explicitly
// listed are also removed, so this resets the scope to the server defaults.
func (c *RustfsAdmin) DeleteConfig(subSystem string) error {
	reqData := RequestData{
		Method:  "DELETE",
		RelPath: "del-config-kv",
		Content: []byte(subSystem),
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	_, err := c.doRequest(ctx, reqData)
	return err
}

// HelpConfig returns the documented keys of a sub-system via GET help-config-kv.
// The returned ConfigKV pairs map each key to its value type hint.
func (c *RustfsAdmin) HelpConfig(subSystem string) ([]ConfigKV, error) {
	query := make(url.Values)
	query.Set("subSys", subSystem)
	reqData := RequestData{
		Method:      "GET",
		RelPath:     "help-config-kv",
		QueryValues: query,
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	resp, err := c.doRequest(ctx, reqData)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var out configHelpResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	kvs := make([]ConfigKV, 0, len(out.KeysHelp))
	for _, entry := range out.KeysHelp {
		kvs = append(kvs, ConfigKV{Key: entry.Key, Value: entry.Type})
	}
	return kvs, nil
}

// buildConfigDirective renders a config scope and its key/value pairs in the
// text format the set-config-kv endpoint accepts and the server renders.
func buildConfigDirective(subSystem string, kvs []ConfigKV) string {
	parts := make([]string, 0, len(kvs)+1)
	parts = append(parts, subSystem)
	for _, kv := range kvs {
		parts = append(parts, kv.Key+`="`+escapeConfigValue(kv.Value)+`"`)
	}
	return strings.Join(parts, " ")
}

// escapeConfigValue mirrors the server-side escaping of config values.
func escapeConfigValue(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, `"`, `\"`)
	return value
}

// parseConfigKVs parses the text/plain body returned by get-config-kv into a
// flat list of key/value pairs. Comment lines (env override hints) are skipped.
func parseConfigKVs(text string) ([]ConfigKV, error) {
	var kvs []ConfigKV
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		lineKVs, err := parseConfigLine(line)
		if err != nil {
			return nil, err
		}
		kvs = append(kvs, lineKVs...)
	}
	return kvs, nil
}

// parseConfigLine parses a single rendered config line of the form
// `scope key="value" key2="value2"` and returns its key/value pairs.
func parseConfigLine(line string) ([]ConfigKV, error) {
	tokens, err := tokenizeConfigLine(line)
	if err != nil {
		return nil, err
	}
	if len(tokens) < 2 {
		return nil, nil
	}
	kvs := make([]ConfigKV, 0, len(tokens)-1)
	for _, token := range tokens[1:] {
		key, value, ok := strings.Cut(token, "=")
		if !ok {
			return nil, fmt.Errorf("config assignment must use key=value syntax: %q", token)
		}
		kvs = append(kvs, ConfigKV{Key: key, Value: value})
	}
	return kvs, nil
}

// tokenizeConfigLine splits a config line into tokens, honoring the quoting and
// backslash escaping rules used by the server's config parser.
func tokenizeConfigLine(line string) ([]string, error) {
	var tokens []string
	var current strings.Builder
	var quote byte
	escaped := false
	for i := 0; i < len(line); i++ {
		ch := line[i]
		if escaped {
			current.WriteByte(ch)
			escaped = false
			continue
		}
		if ch == '\\' {
			escaped = true
			continue
		}
		if quote != 0 {
			if ch == quote {
				quote = 0
			} else {
				current.WriteByte(ch)
			}
			continue
		}
		switch ch {
		case '"', '\'':
			quote = ch
		case ' ', '\t':
			if current.Len() > 0 {
				tokens = append(tokens, current.String())
				current.Reset()
			}
		default:
			current.WriteByte(ch)
		}
	}
	if escaped {
		current.WriteByte('\\')
	}
	if quote != 0 {
		return nil, fmt.Errorf("unterminated quoted config value")
	}
	if current.Len() > 0 {
		tokens = append(tokens, current.String())
	}
	return tokens, nil
}
