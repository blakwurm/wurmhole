package main

import (
	"fmt"
	"runtime/debug"
	"time"

	"github.com/blakwurm/wurmhole/playlist"
	"github.com/gin-gonic/gin"
)

var plist *playlist.DynamicPlaylist
var streams []string
var curStream string
var targetLength float32
var timeToWait float32
var lastUpdate time.Time

const (
	basePlaylistUrl string = "http://localhost:8080/hls/"
	baseKeyframeUrl string = "hls/"
)

func main() {
	streams = make([]string, 0)

	r := gin.Default()

	r.GET("/playlist.m3u8", servePlaylist)
	r.GET("/sources", getSources)
	r.POST("/transition", switchSource)
	r.POST("/stream/begin", streamBegin)
	r.POST("/stream/end", streamEnd)

	r.Run("0.0.0.0:8000")
}

func getPlaylist(stream string) string {
	return fmt.Sprintf("%s%s.m3u8", basePlaylistUrl, stream)
}

func streamBegin(c *gin.Context) {
	defer func() {
		if r := recover(); r != nil {
			str := fmt.Sprintf("%v, %s", r, string(debug.Stack()))
			c.String(500, str)
		}
	}()

	var req PublishRequest
	c.Bind(&req)

	streams = append(streams, req.Name)
	c.Status(200)
}

func streamEnd(c *gin.Context) {
	var req PublishRequest
	c.Bind(&req)
	for i, v := range streams {
		if v == req.Name {
			l := len(streams)
			if l > 1 {
				streams[i] = streams[l-1]
			}
			streams = streams[:l-1]
			break
		}
	}
	c.Status(200)
}

func updatePlaylist() error {
	update, err := plist.UpdateFromUrl(getPlaylist(curStream), false)
	if err != nil {
		return err
	}

	lastUpdate = time.Now()

	if update {
		timeToWait = targetLength
	} else {
		timeToWait = targetLength / 2
	}

	return nil
}

func servePlaylist(c *gin.Context) {
	c.Header("Content-Type", "application/x-mpegURL")
	c.Header("Cache-Control", "no-cache")
	if plist == nil {
		c.String(200, "")
		return
	}

	dur := time.Now().Sub(lastUpdate)
	if float32(dur.Seconds()) >= timeToWait {
		err := updatePlaylist()
		if err != nil {
			msg := fmt.Sprintf("Failed to update playlist\n%v", err)
			c.String(500, msg)
			return
		}
	}

	str := plist.String()

	c.String(200, str)
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

func switchSource(c *gin.Context) {
	var source ContentSource
	c.Bind(&source)

	if contains(streams, source.Name) {
		curStream = source.Name
	} else {
		c.String(400, "Unknown source")
		return
	}

	if plist == nil {
		plistref, err := playlist.NewDynamicPlaylistFromUrl(getPlaylist(curStream), baseKeyframeUrl)
		if err != nil {
			msg := fmt.Sprintf("Failed to create playlist\n%v", err)
			c.String(500, msg)
			return
		}

		plist = plistref

		lastUpdate = time.Now()
		td, err := plist.Header("#EXT-X-TARGETDURATION").Float(32)

		if err != nil {
			msg := fmt.Sprintf("Failed to get target duration\n%v", err)
			c.String(500, msg)
			return
		}

		targetLength = float32(td)
		timeToWait = targetLength
		c.String(200, "OK")
		return
	}

	_, err := plist.UpdateFromUrl(getPlaylist(curStream), true)
	if err != nil {
		msg := fmt.Sprintf("Failed to update playlist\n%v", err)
		c.String(500, msg)
		return
	}

	lastUpdate = time.Now()
	timeToWait = targetLength

	c.String(200, "OK")
}

func getSources(c *gin.Context) {
	c.JSON(200, streams)
}
