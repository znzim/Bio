package lastfm

import (
	"strings"
	"testing"
)

func TestParseTrackUsesFallbackArtworkWhenMissing(t *testing.T) {
	t.Parallel()

	raw := []byte(`{
		"name": "Ghost Song",
		"url": "/music/Kisakay/_/Ghost+Song",
		"artist": { "name": "Kisakay" },
		"image": [
			{ "#text": "" }
		],
		"@attr": { "nowplaying": "true" }
	}`)

	track, err := parseTrack(raw, "Kisakay")
	if err != nil {
		t.Fatalf("parseTrack() error = %v", err)
	}

	if track == nil {
		t.Fatal("parseTrack() returned nil track")
	}

	if track.Artwork == nil {
		t.Fatal("parseTrack() returned nil artwork")
	}

	if !strings.HasPrefix(*track.Artwork, "data:image/svg+xml;base64,") {
		t.Fatalf("expected SVG fallback artwork, got %q", *track.Artwork)
	}
}

func TestParseTrackSkipsLastfmPlaceholderArtwork(t *testing.T) {
	t.Parallel()

	raw := []byte(`{
		"name": "Ghost Song",
		"url": "/music/Kisakay/_/Ghost+Song",
		"artist": { "name": "Kisakay" },
		"image": [
			{ "#text": "https://lastfm.freetls.fastly.net/i/u/64s/2a96cbd8b46e442fc41c2b86b821562f.png" },
			{ "#text": "https://cdn.example.com/cover.jpg" }
		]
	}`)

	track, err := parseTrack(raw, "Kisakay")
	if err != nil {
		t.Fatalf("parseTrack() error = %v", err)
	}

	if track == nil || track.Artwork == nil {
		t.Fatal("parseTrack() returned no artwork")
	}

	if got := *track.Artwork; got != "https://cdn.example.com/cover.jpg" {
		t.Fatalf("expected usable artwork URL, got %q", got)
	}
}
