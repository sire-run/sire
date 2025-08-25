package integration

import ()

// SireAdapter provides access to the Sire core.
type SireAdapter struct{}

// NewSireAdapter creates a new SireAdapter.
func NewSireAdapter() *SireAdapter {
	return &SireAdapter{}
}

// GetNodeTypes returns a list of available node types.
func (a *SireAdapter) GetNodeTypes() []string {
	// This is a placeholder. In a real implementation, this would
	// query the core.globalRegistry. For now, we return a static list.
	return []string{
		"http.request",
		"file.read",
		"file.write",
		"data.transform",
	}
}
