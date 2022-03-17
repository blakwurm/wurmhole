package main

import (
	"strings"
	"testing"

	"github.com/blakwurm/wurmhole/playlist"
)

type StitchTest struct {
	original string
	stitched string
	final    string
	disjoint bool
}

var stitchTests = []StitchTest{
	{
		original: `#EXTINF:10.0,
		http://example.com/1.ts`,

		stitched: `#EXTINF:10.0,
		http://example.com/1.ts
		#EXTINF:10.0,
		http://example.com/2.ts`,

		final: `#EXTINF:10.000,
		http://example.com/1.ts
		#EXTINF:10.000,
		http://example.com/2.ts`,

		disjoint: false,
	},
	{
		original: `#EXTINF:10.0,
		http://example.com/1.ts
		#EXTINF:10.0,
		http://example.com/2.ts
		#EXTINF:10.0,
		http://other.com/1.ts`,

		stitched: `#EXTINF:10.0,
		http://other.com/1.ts
		#EXTINF:10.0,
		http://other.com/2.ts`,

		final: `#EXTINF:10.000,
		http://example.com/1.ts
		#EXTINF:10.000,
		http://example.com/2.ts
		#EXTINF:10.000,
		http://other.com/1.ts
		#EXTINF:10.000,
		http://other.com/2.ts`,

		disjoint: false,
	},
	{
		original: `#EXTINF:10.0,
		http://example.com/1.ts
		#EXTINF:10.0,
		http://example.com/2.ts
		#EXT-X-DISCONTINUITY
		#EXTINF:10.0,
		http://other.com/3.ts
		#EXTINF:10.0,
		http://other.com/4.ts`,

		stitched: `#EXTINF:10.0,
		http://other.com/1.ts
		#EXTINF:10.0,
		http://other.com/2.ts
		#EXTINF:10.0,
		http://other.com/3.ts
		#EXTINF:10.0,
		http://other.com/4.ts`,

		final: `#EXTINF:10.000,
		http://example.com/1.ts
		#EXTINF:10.000,
		http://example.com/2.ts
		#EXT-X-DISCONTINUITY
		#EXTINF:10.000,
		http://other.com/3.ts
		#EXTINF:10.000,
		http://other.com/4.ts`,

		disjoint: false,
	},
	{
		original: `#EXTINF:10.0,
		http://example.com/1.ts
		#EXTINF:10.0,
		http://example.com/2.ts
		#EXT-X-DISCONTINUITY`,

		stitched: `#EXTINF:10.0,
		http://other.com/1.ts
		#EXTINF:10.0,
		http://other.com/2.ts`,

		final: `#EXTINF:10.000,
		http://example.com/1.ts
		#EXTINF:10.000,
		http://example.com/2.ts
		#EXT-X-DISCONTINUITY
		#EXTINF:10.000,
		http://other.com/1.ts
		#EXTINF:10.000,
		http://other.com/2.ts`,

		disjoint: false,
	},
	{
		original: `#EXTINF:10.0,
		http://example.com/1.ts
		#EXTINF:10.0,
		http://example.com/2.ts`,

		stitched: `#EXTINF:10.0,
		http://other.com/1.ts
		#EXTINF:10.0,
		http://other.com/2.ts
		#EXTINF:10.0,
		http://other.com/3.ts`,

		final: `#EXTINF:10.000,
		http://example.com/1.ts
		#EXTINF:10.000,
		http://example.com/2.ts
		#EXT-X-DISCONTINUITY
		#EXTINF:10.000,
		http://other.com/1.ts
		#EXTINF:10.000,
		http://other.com/2.ts
		#EXTINF:10.000,
		http://other.com/3.ts`,

		disjoint: true,
	},
	{
		original: `#EXTINF:10.0,
		http://example.com/1.ts
		#EXTINF:10.0,
		http://example.com/2.ts`,

		stitched: `#EXTINF:10.0,
		http://other.com/1.ts
		#EXT-X-DISCONTINUITY
		#EXTINF:10.0,
		http://other.com/2.ts
		#EXTINF:10.0,
		http://other.com/3.ts`,

		final: `#EXTINF:10.000,
		http://example.com/1.ts
		#EXTINF:10.000,
		http://example.com/2.ts
		#EXT-X-DISCONTINUITY
		#EXTINF:10.000,
		http://other.com/1.ts
		#EXT-X-DISCONTINUITY
		#EXTINF:10.000,
		http://other.com/2.ts
		#EXTINF:10.000,
		http://other.com/3.ts`,

		disjoint: true,
	},
}

func cleanString(s string) string {
	lines := strings.Split(s, "\n")
	var builder strings.Builder

	for i, line := range lines {
		if i > 0 {
			builder.WriteString("\n")
		}
		builder.WriteString(strings.TrimSpace(line))
	}

	return builder.String()
}

func prepareTests() {
	for i, test := range stitchTests {
		stitchTests[i].original = cleanString(test.original)
		stitchTests[i].stitched = cleanString(test.stitched)
		stitchTests[i].final = cleanString(test.final)
	}
}

func TestStitch(t *testing.T) {
	prepareTests()
	prefix := `#EXTM3U
#EXT-X-MEDIA-SEQUENCE:`
	prefix_original := prefix + "1\n"
	prefix_stitched := prefix + "2\n"

	for i, test := range stitchTests {
		original := prefix_original + test.original + "\n"
		stitched := prefix_stitched + test.stitched + "\n"
		final := prefix + "2\n" + test.final + "\n"

		oplist, err := playlist.ParsePlaylist(strings.NewReader(original))
		if err != nil {
			t.Errorf("Test %d: %v", i, err)
			continue
		}

		stlist, err := playlist.ParsePlaylist(strings.NewReader(stitched))
		if err != nil {
			t.Errorf("Test %d: %v", i, err)
			continue
		}

		p, err := playlist.NewDynamicPlaylist(oplist, "", true)
		if err != nil {
			t.Errorf("Test %d: %s", i, err)
			continue
		}

		_, err = p.UpdatePlaylist(stlist, test.disjoint)
		if err != nil {
			t.Errorf("Test %d: %s", i, err)
			continue
		}

		if p.String() != final {
			t.Errorf("\nTest %d: Expected:\n%s\ngot:\n%s", i, final, p.String())
		}
	}
}

func BenchmarkDynamicPlaylist(b *testing.B) {
	prefix := `#EXTM3U
#EXT-X-MEDIA-SEQUENCE:`
	prefix_original := prefix + "1\n"
	prefix_stitched := prefix + "2\n"

	for _, test := range stitchTests {
		original := prefix_original + test.original + "\n"
		stitched := prefix_stitched + test.stitched + "\n"

		oplist, _ := playlist.ParsePlaylist(strings.NewReader(original))
		stlist, _ := playlist.ParsePlaylist(strings.NewReader(stitched))
		p, _ := playlist.NewDynamicPlaylist(oplist, "", true)
		_, _ = p.UpdatePlaylist(stlist, test.disjoint)
	}
}
