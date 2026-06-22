package theme

import (
	"sort"

	"github.com/GetStoreConnect/storeconnect-cli/internal/api"
)

// CategoryDiff holds the keys that differ between local and server for one
// category of theme content (templates or assets). Added are present locally
// but not on the server; Removed are present on the server but not locally;
// Modified are present in both but with differing content. All slices are
// sorted for stable output.
type CategoryDiff struct {
	Added    []string `json:"added"`
	Modified []string `json:"modified"`
	Removed  []string `json:"removed"`
}

// count returns the total number of differing keys in this category.
func (c CategoryDiff) count() int {
	return len(c.Added) + len(c.Modified) + len(c.Removed)
}

// Diff is the full divergence between a local theme and the server's copy.
type Diff struct {
	Templates CategoryDiff `json:"templates"`
	Assets    CategoryDiff `json:"assets"`
}

// HasChanges reports whether the local theme differs from the server in any way.
func (d Diff) HasChanges() bool {
	return d.Templates.count() > 0 || d.Assets.count() > 0
}

// PushableCount is the number of keys that `sc theme push` would send: added
// and modified templates and assets. Removed keys (present on the server but
// not locally) are excluded because push never deletes server content.
func (d Diff) PushableCount() int {
	return len(d.Templates.Added) + len(d.Templates.Modified) +
		len(d.Assets.Added) + len(d.Assets.Modified)
}

// ServerOnlyCount is the number of keys present on the server but not locally
// (removed templates and assets) — divergence that push will not reconcile.
func (d Diff) ServerOnlyCount() int {
	return len(d.Templates.Removed) + len(d.Assets.Removed)
}

// ComputeDiff compares a local theme against the server's copy. Templates are
// compared by key, treating a content difference as a modification. Assets are
// compared by key using their SHA-256 content hash, so an unchanged binary is
// never reported even though its bytes are read locally. ComputeDiff is pure so
// the comparison rule is trivially unit-testable.
func ComputeDiff(local *api.Theme, localAssets []LocalAsset, server *api.Theme) Diff {
	return Diff{
		Templates: diffTemplates(local, server),
		Assets:    diffAssets(localAssets, server),
	}
}

func diffTemplates(local, server *api.Theme) CategoryDiff {
	localByKey := make(map[string]string, len(local.Templates))
	for _, t := range local.Templates {
		localByKey[t.Key] = t.Content
	}
	serverByKey := make(map[string]string, len(server.Templates))
	for _, t := range server.Templates {
		serverByKey[t.Key] = t.Content
	}

	var diff CategoryDiff
	for key, localContent := range localByKey {
		serverContent, exists := serverByKey[key]
		switch {
		case !exists:
			diff.Added = append(diff.Added, key)
		case localContent != serverContent:
			diff.Modified = append(diff.Modified, key)
		}
	}
	for key := range serverByKey {
		if _, exists := localByKey[key]; !exists {
			diff.Removed = append(diff.Removed, key)
		}
	}

	sortCategory(&diff)
	return diff
}

func diffAssets(local []LocalAsset, server *api.Theme) CategoryDiff {
	localByKey := make(map[string]string, len(local))
	for _, a := range local {
		localByKey[a.Key] = a.ContentHash
	}
	serverByKey := make(map[string]string, len(server.Assets))
	for _, a := range server.Assets {
		serverByKey[a.Key] = a.ContentHash
	}

	var diff CategoryDiff
	for key, localHash := range localByKey {
		serverHash, exists := serverByKey[key]
		switch {
		case !exists:
			diff.Added = append(diff.Added, key)
		case localHash != serverHash:
			diff.Modified = append(diff.Modified, key)
		}
	}
	for key := range serverByKey {
		if _, exists := localByKey[key]; !exists {
			diff.Removed = append(diff.Removed, key)
		}
	}

	sortCategory(&diff)
	return diff
}

func sortCategory(c *CategoryDiff) {
	sort.Strings(c.Added)
	sort.Strings(c.Modified)
	sort.Strings(c.Removed)
}
