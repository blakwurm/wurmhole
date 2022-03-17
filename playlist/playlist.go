package playlist

import (
	"bufio"
	"errors"
	"io"
	"strconv"
	"strings"
)

type Entry struct {
	length   float32
	location string
	disjoint bool
}

func Disjoint() Entry {
	return Entry{
		disjoint: true,
	}
}

func (e *Entry) String(pathFixers ...func(str string) string) string {
	if e.disjoint {
		return "#EXT-X-DISCONTINUITY"
	}
	str := e.location

	for _, fixer := range pathFixers {
		str = fixer(str)
	}

	return "#EXTINF:" + strconv.FormatFloat(float64(e.length), 'f', 3, 32) + ",\n" + str
}

type Playlist struct {
	headers map[string]HeaderValue
	entries []Entry
}

func EmptyPlaylist() Playlist {
	return Playlist{
		headers: make(map[string]HeaderValue),
		entries: make([]Entry, 0),
	}
}

func ParsePlaylist(source io.Reader) (Playlist, error) {
	scanner := bufio.NewScanner(source)
	if !scanner.Scan() {
		return Playlist{}, errors.New("No format header")
	}

	head := scanner.Text()
	if !strings.HasPrefix(head, "#EXTM3U") {
		return Playlist{}, errors.New("Not a M3U valid playlist")
	}

	playlist := Playlist{
		headers: make(map[string]HeaderValue),
		entries: make([]Entry, 0),
	}

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "#EXT-X-ENDLIST" {
			break
		}

		if line == "#EXT-X-DISCONTINUITY" {
			playlist.entries = append(playlist.entries, Disjoint())
			continue
		}

		if line[len(line)-1] != ',' {
			key, value := ParseHeader(line)
			playlist.headers[key] = value
			continue
		}

		split := strings.SplitN(line, ":", 2)

		numStr := split[1][:len(split[1])-1]
		length, err := strconv.ParseFloat(numStr, 32)

		if err != nil {
			return Playlist{}, errors.New("Malformed playlist")
		}

		if !scanner.Scan() {
			return Playlist{}, errors.New("Malformed playlist")
		}

		payload := strings.TrimSpace(scanner.Text())
		entry := Entry{
			length:   float32(length),
			location: payload,
			disjoint: false,
		}

		playlist.entries = append(playlist.entries, entry)
	}

	return playlist, nil
}

func (p *Playlist) Header(key string) HeaderValue {
	return p.headers[key]
}

func (p *Playlist) Headers() map[string]HeaderValue {
	return p.headers
}

func (p *Playlist) Entries() []Entry {
	return p.entries
}

func (p *Playlist) LatestEntry() Entry {
	return p.entries[len(p.entries)-1]
}

func (p *Playlist) String(pathFixers ...func(str string) string) string {
	var builder strings.Builder

	builder.WriteString("#EXTM3U\n")

	for k, v := range p.headers {
		builder.WriteString("#")
		builder.WriteString(k)
		builder.WriteString(":")
		builder.WriteString(v.String())
		builder.WriteString("\n")
	}

	for _, e := range p.entries {
		builder.WriteString(e.String(pathFixers...))
		builder.WriteString("\n")
	}

	builder.WriteString("#EXT-X-ENDLIST\n")

	return builder.String()
}
