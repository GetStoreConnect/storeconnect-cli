package theme

import (
	"reflect"
	"testing"

	"github.com/GetStoreConnect/storeconnect-cli/internal/api"
)

func tmpl(key, content string) api.ThemeTemplate {
	return api.ThemeTemplate{Key: key, Content: content}
}

func TestComputeDiffTemplates(t *testing.T) {
	local := &api.Theme{Templates: []api.ThemeTemplate{
		tmpl("layouts/theme", "CHANGED"), // modified
		tmpl("pages/home", "same"),       // unchanged
		tmpl("pages/new", "brand new"),   // added
	}}
	server := &api.Theme{Templates: []api.ThemeTemplate{
		tmpl("layouts/theme", "original"),
		tmpl("pages/home", "same"),
		tmpl("pages/gone", "old"), // removed
	}}

	diff := ComputeDiff(local, nil, server)

	if !reflect.DeepEqual(diff.Templates.Added, []string{"pages/new"}) {
		t.Errorf("added = %v", diff.Templates.Added)
	}
	if !reflect.DeepEqual(diff.Templates.Modified, []string{"layouts/theme"}) {
		t.Errorf("modified = %v", diff.Templates.Modified)
	}
	if !reflect.DeepEqual(diff.Templates.Removed, []string{"pages/gone"}) {
		t.Errorf("removed = %v", diff.Templates.Removed)
	}
}

func TestComputeDiffAssets(t *testing.T) {
	local := []LocalAsset{
		{Key: "logo.png", ContentHash: "newhash"},      // modified
		{Key: "fonts/brand.woff2", ContentHash: "abc"}, // added
		{Key: "icon.svg", ContentHash: "same"},         // unchanged
	}
	server := &api.Theme{Assets: []api.ThemeAsset{
		{Key: "logo.png", ContentHash: "oldhash"},
		{Key: "icon.svg", ContentHash: "same"},
		{Key: "legacy.gif", ContentHash: "x"}, // removed
	}}

	diff := ComputeDiff(&api.Theme{}, local, server)

	if !reflect.DeepEqual(diff.Assets.Added, []string{"fonts/brand.woff2"}) {
		t.Errorf("added = %v", diff.Assets.Added)
	}
	if !reflect.DeepEqual(diff.Assets.Modified, []string{"logo.png"}) {
		t.Errorf("modified = %v", diff.Assets.Modified)
	}
	if !reflect.DeepEqual(diff.Assets.Removed, []string{"legacy.gif"}) {
		t.Errorf("removed = %v", diff.Assets.Removed)
	}
}

func TestComputeDiffInSync(t *testing.T) {
	local := &api.Theme{Templates: []api.ThemeTemplate{tmpl("layouts/theme", "x")}}
	localAssets := []LocalAsset{{Key: "logo.png", ContentHash: "h"}}
	server := &api.Theme{
		Templates: []api.ThemeTemplate{tmpl("layouts/theme", "x")},
		Assets:    []api.ThemeAsset{{Key: "logo.png", ContentHash: "h"}},
	}

	diff := ComputeDiff(local, localAssets, server)

	if diff.HasChanges() {
		t.Errorf("expected no changes, got %+v", diff)
	}
	if diff.PushableCount() != 0 || diff.ServerOnlyCount() != 0 {
		t.Errorf("expected zero counts, got pushable=%d serverOnly=%d", diff.PushableCount(), diff.ServerOnlyCount())
	}
}

func TestComputeDiffCounts(t *testing.T) {
	local := &api.Theme{Templates: []api.ThemeTemplate{
		tmpl("a", "1"),       // added
		tmpl("b", "changed"), // modified
	}}
	localAssets := []LocalAsset{{Key: "c.png", ContentHash: "new"}} // added
	server := &api.Theme{
		Templates: []api.ThemeTemplate{
			tmpl("b", "orig"),
			tmpl("d", "x"), // removed
		},
		Assets: []api.ThemeAsset{{Key: "e.png", ContentHash: "y"}}, // removed
	}

	diff := ComputeDiff(local, localAssets, server)

	if !diff.HasChanges() {
		t.Fatal("expected changes")
	}
	// added a, modified b, added c.png = 3 pushable
	if diff.PushableCount() != 3 {
		t.Errorf("pushable = %d, want 3", diff.PushableCount())
	}
	// removed d, removed e.png = 2 server-only
	if diff.ServerOnlyCount() != 2 {
		t.Errorf("serverOnly = %d, want 2", diff.ServerOnlyCount())
	}
}

func TestComputeDiffEmptyServer(t *testing.T) {
	local := &api.Theme{Templates: []api.ThemeTemplate{tmpl("a", "1")}}
	localAssets := []LocalAsset{{Key: "b.png", ContentHash: "h"}}

	diff := ComputeDiff(local, localAssets, &api.Theme{})

	if !reflect.DeepEqual(diff.Templates.Added, []string{"a"}) {
		t.Errorf("templates added = %v", diff.Templates.Added)
	}
	if !reflect.DeepEqual(diff.Assets.Added, []string{"b.png"}) {
		t.Errorf("assets added = %v", diff.Assets.Added)
	}
	if diff.ServerOnlyCount() != 0 {
		t.Errorf("serverOnly = %d, want 0", diff.ServerOnlyCount())
	}
}
