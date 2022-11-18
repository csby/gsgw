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

	return instance
}

type Handler struct {
	gtype.Base

	cloud gcloud.Handler
	node  gnode.Handler
}

func (s *Handler) InitRouting(router gtype.Router) {
}

func (s *Handler) BeforeRouting(ctx gtype.Context) {
	method := ctx.Method()

	// enable across access
	if method == http.MethodOptions {
		ctx.Response().Header().Add("Access-Control-Allow-Origin", "*")
		ctx.Response().Header().Set("Access-Control-Allow-Headers", "content-type,token")
		ctx.SetHandled(true)
		return
	}

	if method == http.MethodGet {
		schema := ctx.Schema()
		path := ctx.Path()

		// default to opt site
		if "/" == path || "" == path || gopt.WebPath == path {
			redirectUrl := fmt.Sprintf("%s://%s%s/", schema, ctx.Host(), gopt.WebPath)
			http.Redirect(ctx.Response(), ctx.Request(), redirectUrl, http.StatusMovedPermanently)
			ctx.SetHandled(true)
			return
		}

		// http to https
		if "http" == method {
			if cfg.Http.RedirectToHttps && cfg.Https.Enabled {
				redirectUrl := fmt.Sprintf("%s://%s%s", "https", ctx.Host(), path)
				http.Redirect(ctx.Response(), ctx.Request(), redirectUrl, http.StatusMovedPermanently)
				ctx.SetHandled(true)
				return
			}
		}
	}
}

func (s *Handler) AfterRouting(ctx gtype.Context) {

}

func (s *Handler) AfterInput(ctx gtype.Context) {

}

func (s *Handler) ExtendOptSetup(opt gtype.Option) {
	if opt == nil {
		return
	}

	opt.SetCloud(cfg.Cloud.Enabled)
	opt.SetNode(cfg.Node.Enabled)
}

func (s *Handler) ExtendOptApi(router gtype.Router,
	path *gtype.Path,
	preHandle gtype.HttpHandle,
	opt gtype.Opt) {
	s.cloud = gcloud.NewHandler(s.GetLog(), &cfg.Config, opt.Wsc())
	s.node = gnode.NewHandler(s.GetLog(), &cfg.Config, opt.Wsc())

	if cfg.Cloud.Enabled {
		s.cloud.Init(router, path, preHandle, nil)
	}
	if cfg.Node.Enabled {
		s.node.Init(router, path, preHandle)
	}
}
