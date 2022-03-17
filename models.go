package main

type ContentSource struct {
	Name string `form:"name"`
}

type PublishRequest struct {
	Name string `form:"name"`
}
