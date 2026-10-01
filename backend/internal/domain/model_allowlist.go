package domain

// GroupModelAllowlist is the persisted group-level model allowlist.
// When enabled it constrains both model discovery and gateway admission.
type GroupModelAllowlist struct {
	Enabled bool     `json:"enabled"`
	Models  []string `json:"models,omitempty"`
}
