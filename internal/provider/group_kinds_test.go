package provider

import (
	"slices"
	"testing"
)

// A group takes every kind of input a member's source listed, not images
// alone: a routing group of a model that takes video (Kimi K3's own list
// says text, image and video) was told text and image, so an agent was
// never offered video for it.
//
// buildGroupEntries aggregates Images from its members, but Inputs and
// Outputs were never read, and Modalities falls back to text and image when
// Inputs is nil.
func TestGroupKindsFromMembers(t *testing.T) {
	for name, c := range map[string]struct {
		in   []string
		in2  []string
		want []string
	}{
		"one takes video":  {[]string{"text", "image", "video"}, nil, []string{"text", "image", "video"}},
		"neither said":     {nil, nil, nil},
		"both say text":    {[]string{"text", "image"}, []string{"text", "image"}, []string{"text", "image"}},
		"union in order":   {[]string{"text", "video"}, []string{"text", "image", "pdf"}, []string{"text", "video", "image", "pdf"}},
		"audio from other": {[]string{"text"}, []string{"text", "audio"}, []string{"text", "audio"}},
	} {
		a := quiet("a")
		a.Inputs = c.in
		b := quiet("b")
		b.Inputs = c.in2
		g := groupEntries([]Entry{a, b})
		if len(g) != 1 {
			t.Fatalf("%s: groups %+v", name, g)
		}
		if !slices.Equal(g[0].Inputs, c.want) {
			t.Errorf("%s: group Inputs %v, want %v", name, g[0].Inputs, c.want)
		}
	}
}

// The kinds a group is told reach an agent: Modalities reads them beside
// Images, so a group of a model that takes video says so.
func TestGroupModalitiesIncludeVideo(t *testing.T) {
	a := quiet("a")
	a.Images = true
	a.Inputs = []string{"text", "image", "video"}
	b := quiet("b")
	b.Inputs = []string{"text", "image", "video"}
	g := groupEntries([]Entry{a, b})
	if len(g) != 1 {
		t.Fatalf("groups %+v", g)
	}
	in, _ := g[0].Modalities()
	want := []string{"text", "image", "video"}
	if !slices.Equal(in, want) {
		t.Errorf("group modalities %v, want %v", in, want)
	}
}

// A group whose members say nothing of their kinds keeps saying nothing:
// nil is Modalities' fallback to text and image, an empty list is not.
func TestGroupKindsStayUnknown(t *testing.T) {
	g := groupEntries([]Entry{quiet("a"), quiet("b")})
	if len(g) != 1 {
		t.Fatalf("groups %+v", g)
	}
	if g[0].Inputs != nil || g[0].Outputs != nil {
		t.Errorf("group invented kinds: in %v out %v", g[0].Inputs, g[0].Outputs)
	}
}

// Outputs, too: a group said to give only text when a member says so.
func TestGroupOutputKinds(t *testing.T) {
	a := quiet("a")
	a.Outputs = []string{"text"}
	b := quiet("b")
	b.Outputs = []string{"text"}
	if g := groupEntries([]Entry{a, b}); len(g) != 1 || !slices.Equal(g[0].Outputs, []string{"text"}) {
		t.Fatalf("group Outputs: %+v", g)
	}
}
