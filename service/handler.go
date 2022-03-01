package main

import (
	"fmt"
	"github.com/csby/gwsf/gcloud"
	"github.com/csby/gwsf/gnode"
	"github.com/csby/gwsf/gopt"
	"github.com/csby/gwsf/gtype"
	"net/http"
)

func NewHandler(log gtype.Log) gtype.Handler {
	instance := &Handler{}
	instance.SetLog(log)

	instance.apiController = &Controller{}
	instance.apiController.SetLog(log)

	return instance
}

type Handler struct {
	gtype.Base

	cloud gcloud.Handler
	node  gnode.Handler

	apiController *Controller
}

func (s *Handler) InitRouting(router gtype.Router) {
}

func (s *Handler) BeforeRouting(ctx gtype.Context) {
	method := ctx.Method()

	// enable across access
	if method == "OPTIONS" {
		ctx.Response().Header().Add("Access-Control-Allow-Origin", "*")
		ctx.Response().Header().Set("Access-Control-Allow-Headers", "content-type,token")
		ctx.SetHandled(true)
		return
	}

	// default to opt site
	if method == "GET" {
		path := ctx.Path()
		if "/" == path || "" == path || gopt.WebPath == path {
			redirectUrl := fmt.Sprintf("%s://%s%s/", ctx.Schema(), ctx.Host(), gopt.WebPath)
			http.Redirect(ctx.Response(), ctx.Request(), redirectUrl, http.StatusMovedPermanently)
			ctx.SetHandled(true)
			return
		}
	}
}

func (s *Handler) AfterRouting(ctx gtype.Context) {

}

func (s *Handler) ExtendOptSetup(opt gtype.Option) {
	if opt == nil {
		return
	}

	opt.SetCloud(cfg.Cloud.Enabled)
	opt.SetNode(cfg.Node.Enabled)
}

func (s *Handler) ExtendOptApi(router gtype.Router, path *gtype.Path, preHandle gtype.HttpHandle, wsc gtype.SocketChannelCollection) {
	s.cloud = gcloud.NewHandler(s.GetLog(), &cfg.Config, wsc)
	s.node = gnode.NewHandler(s.GetLog(), &cfg.Config, wsc)

	s.cloud.Init(router, path, preHandle, nil)
	s.node.Init(router, path, preHandle)
}
