package agent

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yetone/magpie/internal/catalog"
	"github.com/yetone/magpie/internal/edit"
	"github.com/yetone/magpie/internal/provider"
)

// #290: a kimi-cli user who moved to the new Kimi Code has both ~/.kimi and
// ~/.kimi-code; the new one reads only ~/.kimi-code, so that is where the
// model picked for it goes — ~/.kimi is left as it was.
func TestKimiCode(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))
	t.Setenv("XDG_CACHE_HOME", filepath.Join(home, ".cache"))
	t.Setenv("KIMI_SHARE_DIR", "")
	t.Setenv("KIMI_CODE_HOME", "")
	if err := provider.Save(provider.Provider{ID: "relay", Name: "Relay", Chat: "https://example.com/v1", Key: "k", Models: []string{"gemini-3.8-flash"}}); err != nil {
		t.Fatal(err)
	}
	cfg := `default_model = "kimi-code/kimi-for-coding"

[providers."managed:kimi-code"]
type = "kimi"
base_url = "https://api.kimi.com/coding/v1"
api_key = ""

[models."kimi-code/kimi-for-coding"]
provider = "managed:kimi-code"
model = "kimi-for-coding"
max_context_size = 262144
`
	old := filepath.Join(home, ".kimi", "config.toml")
	path := filepath.Join(home, ".kimi-code", "config.toml")
	for _, p := range []string{old, path} {
		os.MkdirAll(filepath.Dir(p), 0o755)
		os.WriteFile(p, []byte(cfg), 0o644)
	}
	a := kimi(home)
	if a.Path != path {
		t.Fatalf("path: %s, want %s", a.Path, path)
	}
	f := a.Field("model")
	if err := f.Set("magpie/relay/gemini-3.8-flash"); err != nil {
		t.Fatal(err)
	}
	if v, _ := edit.GetTOMLTop(path, "default_model"); v != "magpie/relay/gemini-3.8-flash" {
		t.Fatalf("default_model in ~/.kimi-code: %q", v)
	}
	m, err := edit.GetTOMLTable(path, `models."magpie/relay/gemini-3.8-flash"`)
	if err != nil || m == nil || m["provider"] != "magpie" || !kimiToolUse(path) {
		t.Fatalf("model table: %v %v", m, err)
	}
	if b, _ := os.ReadFile(old); string(b) != cfg {
		t.Fatalf("~/.kimi changed:\n%s", b)
	}
	if err := f.Set(""); err != nil {
		t.Fatal(err)
	}
	if b, _ := os.ReadFile(path); strings.Contains(string(b), "magpie") || !strings.Contains(string(b), `default_model = "kimi-code/kimi-for-coding"`) {
		t.Fatalf("reset:\n%s", b)
	}

	// $KIMI_CODE_HOME moves it; kimi-cli's alone is still kimi-cli's, and
	// with neither there the new one's is where a first start puts it
	t.Setenv("KIMI_CODE_HOME", filepath.Join(home, "kc"))
	if d, legacy := KimiDir(home); d != filepath.Join(home, "kc") || legacy {
		t.Fatalf("KIMI_CODE_HOME: %s %v", d, legacy)
	}
	t.Setenv("KIMI_CODE_HOME", "")
	os.RemoveAll(filepath.Join(home, ".kimi-code"))
	if d, legacy := KimiDir(home); d != filepath.Join(home, ".kimi") || !legacy {
		t.Fatalf("kimi-cli only: %s %v", d, legacy)
	}
	// kimi-cli refuses a capability it doesn't know
	if err := kimi(home).Field("model").Set("magpie/relay/gemini-3.8-flash"); err != nil {
		t.Fatal(err)
	}
	if m, _ := edit.GetTOMLTable(old, `models."magpie/relay/gemini-3.8-flash"`); m == nil || kimiToolUse(old) {
		t.Fatalf("kimi-cli's model table: %v", m)
	}
	os.RemoveAll(filepath.Join(home, ".kimi"))
	if d, legacy := KimiDir(home); d != filepath.Join(home, ".kimi-code") || legacy {
		t.Fatalf("neither: %s %v", d, legacy)
	}
}

// kimiToolUse is whether a model table in the file says it calls tools.
func kimiToolUse(path string) bool {
	b, _ := os.ReadFile(path)
	return strings.Contains(string(b), `"tool_use"`)
}

// Kimi Code sets the API for a whole provider, so a model's native wire is
// picked on its own provider table. A sync must keep all three wires rather
// than putting every model back on Chat.
func TestKimiCodeKeepsNativeProtocols(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("XDG_CONFIG_HOME", filepath.Join(home, ".config"))
	t.Setenv("XDG_CACHE_HOME", filepath.Join(home, ".cache"))
	t.Setenv("KIMI_SHARE_DIR", "")
	t.Setenv("KIMI_CODE_HOME", "")
	if err := provider.Save(provider.Provider{
		ID: "chat", Name: "Chat", Chat: "https://chat.example/v1", Key: "k", Models: []string{"glm"},
	}); err != nil {
		t.Fatal(err)
	}
	if err := provider.Save(provider.Provider{
		ID: "resp", Name: "Responses", Responses: "https://resp.example/v1", Key: "k", Models: []string{"gpt-5.5", "mixed"},
	}); err != nil {
		t.Fatal(err)
	}
	if err := provider.Save(provider.Provider{
		ID: "anth", Name: "Messages", Anthropic: "https://anth.example", Key: "k", Models: []string{"claude-opus-5-5"},
	}); err != nil {
		t.Fatal(err)
	}
	if err := catalog.SaveLive("chat", "https://chat.example/v1", []catalog.Model{{ID: "glm", APIs: []string{"chat"}}}); err != nil {
		t.Fatal(err)
	}
	if err := catalog.SaveLive("resp", "https://resp.example/v1", []catalog.Model{
		{ID: "gpt-5.5", APIs: []string{"responses"}},
		{ID: "mixed", APIs: []string{"responses"}},
	}); err != nil {
		t.Fatal(err)
	}
	if err := catalog.SaveLive("anth", "https://anth.example", []catalog.Model{{ID: "claude-opus-5-5", APIs: []string{"anthropic"}}}); err != nil {
		t.Fatal(err)
	}
	for _, g := range []provider.Group{
		{ID: "chat", Name: "Chat", Members: []string{"chat/glm"}},
		{ID: "resp", Name: "Responses", Members: []string{"resp/gpt-5.5"}},
		{ID: "anth", Name: "Messages", Members: []string{"anth/claude-opus-5-5"}},
		{ID: "mixed", Name: "Mixed", Members: []string{"resp/mixed", "chat/glm"}},
	} {
		if err := provider.SaveGroup(g); err != nil {
			t.Fatal(err)
		}
	}
	path := filepath.Join(home, ".kimi-code", "config.toml")
	writeFile(t, path, "")
	a := kimi(home)
	if err := a.Field("model").Set("magpie/group/resp"); err != nil {
		t.Fatal(err)
	}
	table := func(name string) map[string]string {
		t.Helper()
		got, err := edit.GetTOMLTable(path, name)
		if err != nil {
			t.Fatal(err)
		}
		return got
	}
	want := map[string]string{
		"chat/glm":             "magpie",
		"group/chat":           "magpie",
		"resp/gpt-5.5":         "magpie-responses",
		"group/resp":           "magpie-responses",
		"anth/claude-opus-5-5": "magpie-anthropic",
		"group/anth":           "magpie-anthropic",
		"group/mixed":          "magpie",
	}
	check := func(when string) {
		t.Helper()
		for model, providerID := range want {
			got := table(`models."magpie/` + model + `"`)
			if got == nil || got["provider"] != providerID {
				t.Errorf("%s %s: provider %q, want %q", when, model, got["provider"], providerID)
			}
		}
		for tableName, typ := range map[string]string{
			"providers.magpie":           "kimi",
			"providers.magpie-responses": "openai_responses",
			"providers.magpie-anthropic": "anthropic",
		} {
			got := table(tableName)
			if got == nil || got["type"] != typ {
				t.Errorf("%s: %v, want type %q", tableName, got, typ)
			}
		}
		if got := table(`models."magpie/group/anth"`); got["adaptive_thinking"] != "true" {
			t.Error("Messages group has no adaptive_thinking")
		}
	}
	check("pick")
	if err := edit.DelTOMLTable(path, "providers.magpie-responses"); err != nil {
		t.Fatal(err)
	}
	if err := edit.DelTOMLTable(path, "providers.magpie-anthropic"); err != nil {
		t.Fatal(err)
	}
	if err := a.Sync(); err != nil {
		t.Fatal(err)
	}
	check("sync")
	if why := a.Check(); why != "" {
		t.Errorf("Check after sync: %s", why)
	}
	if err := edit.SetTOMLKey(path, "providers.magpie-anthropic", "base_url", "https://wrong.example"); err != nil {
		t.Fatal(err)
	}
	if why := a.Check(); !strings.Contains(why, "providers.magpie-anthropic") {
		t.Errorf("Check missed a drifted Messages provider: %q", why)
	}
}
