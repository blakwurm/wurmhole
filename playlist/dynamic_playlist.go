package playlist

import (
	"math"
	"net/http"
	"strconv"
	"strings"
)

type DynamicPlaylist struct {
	Playlist
	lastCount         int
	mediaSequence     int
	prefix            string
	discontinuityTags bool
}

func dupPlaylist(source Playlist) Playlist {
	playlist := Playlist{
		headers: make(map[string]HeaderValue),
		entries: make([]Entry, 0),
	}

	for k, v := range source.headers {
		playlist.headers[k] = v
	}

	for _, e := range source.entries {
		playlist.entries = append(playlist.entries, e)
	}

	return playlist
}

func downloadPlaylist(sourceUrl string) (Playlist, error) {
	resp, err := http.Get(sourceUrl)
	if err != nil {
		return Playlist{}, err
	}
	body := resp.Body
	defer body.Close()

	return ParsePlaylist(body)
}

func NewDynamicPlaylistFromUrl(sourceUrl string, baseUrl string, discontinuityTags bool) (*DynamicPlaylist, error) {
	source, err := downloadPlaylist(sourceUrl)
	if err != nil {
		return nil, err
	}
	p, err := NewDynamicPlaylist(source, baseUrl, discontinuityTags)
	return p, err
}

func NewDynamicPlaylist(source Playlist, prefix string, discontinuityTags bool) (*DynamicPlaylist, error) {

	return &DynamicPlaylist{
		Playlist:          dupPlaylist(source),
		lastCount:         0,
		mediaSequence:     1,
		prefix:            prefix,
		discontinuityTags: discontinuityTags,
	}, nil
}

func (p *DynamicPlaylist) incrementMediaSequence() {
	p.mediaSequence++
	p.headers["EXT-X-MEDIA-SEQUENCE"] = ToHeaderValue(strconv.Itoa(p.mediaSequence))
}

func (p *DynamicPlaylist) stitchNew(source Playlist) {
	length := len(source.entries)
	search := int(math.Max(float64(length-4), 0))
	stitch := make([]Entry, 0)
	for i := length - 1; float64(i) >= math.Max(float64(search), 0); i-- {
		stitch = append([]Entry{source.entries[i]}, stitch...)
		if source.entries[i].disjoint {
			search--
		}
	}

	if p.discontinuityTags {
		stitch = append([]Entry{Disjoint()}, stitch...)
	}

	p.entries = append(p.entries, stitch...)
}

func (p *DynamicPlaylist) stitch(source Playlist) {
	i := len(source.entries) - 1
	for ; i >= 0; i-- {
		if source.entries[i].disjoint {
			continue
		}
		if p.entries[len(p.entries)-1].location == source.entries[i].location {
			i++
			break
		}
	}
	if i < 0 {
		i++
	}
	p.entries = append(p.entries, source.entries[i:]...)
	pLen := len(p.entries)
	if pLen > 12 {
		p.entries = p.entries[pLen-12:]
	}
}

func (p *DynamicPlaylist) UpdateFromUrl(sourceUrl string, disjoint bool) (bool, error) {
	source, err := downloadPlaylist(sourceUrl)
	if err != nil {
		return false, err
	}

	return p.UpdatePlaylist(source, disjoint)
}

func (p *DynamicPlaylist) UpdatePlaylist(source Playlist, disjoint bool) (bool, error) {
	if disjoint {
		p.stitchNew(source)
		p.incrementMediaSequence()

		return true, nil
	}

	seq, err := source.Header("EXT-X-MEDIA-SEQUENCE").Int(32)
	if err != nil {
		return false, err
	}

	if p.lastCount != int(seq) {
		p.stitch(source)
		p.incrementMediaSequence()
		return true, nil
	}

	return false, nil
}

func (p *DynamicPlaylist) String() string {
	str := p.Playlist.String(func(str string) string {
		return p.prefix + str
	})
	return strings.Replace(str, "#EXT-X-ENDLIST\n", "", 1)
}
