package context

import (
	"github.com/FISTOFDARKNESS/Asset-Reuploader/internal/app/response"
	"github.com/FISTOFDARKNESS/Asset-Reuploader/internal/roblox"
)

type Context struct {
	Client          *roblox.Client
	Logger          *logger
	PauseController *pauseController
	Response        *response.Response
}

func New(c *roblox.Client, resp *response.Response) *Context {
	return &Context{
		Client:          c,
		Logger:          newLogger(),
		PauseController: newPauseController(),
		Response:        resp,
	}
}
