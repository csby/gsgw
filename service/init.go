package main

import (
	"fmt"
	"github.com/csby/gsgw/config"
	"github.com/csby/gwsf/glog"
	"github.com/csby/gwsf/gserver"
	"github.com/csby/gwsf/gtype"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	moduleType    = "server"
	moduleName    = "gsgw"
	moduleRemark  = "安全网关"
	moduleVersion = "1.0.4.1"
)

var (
	cfg              = config.NewConfig()
	log              = &glog.Writer{Level: glog.LevelAll}
	svr gtype.Server = nil
)

func init() {
	moduleArgs := &gtype.Args{}
	serverArgs := &gtype.SvcArgs{}
	moduleArgs.Parse(os.Args, moduleType, moduleName, moduleVersion, moduleRemark, serverArgs)
	now := time.Now()
	cfg.Module.Type = moduleType
	cfg.Module.Name = moduleName
	cfg.Module.Version = moduleVersion
	cfg.Module.Remark = moduleRemark
	cfg.Module.Path = moduleArgs.ModulePath()
	cfg.Svc.BootTime = now
	cfg.Node.InstanceId = gtype.NewGuid()

	rootFolder := filepath.Dir(moduleArgs.ModuleFolder())
	cfgFolder := filepath.Join(rootFolder, "cfg")
	cfgName := fmt.Sprintf("%s.json", moduleName)
	if serverArgs.Help {
		serverArgs.ShowHelp(cfgFolder, cfgName)
		os.Exit(11)
	}

	if serverArgs.Pkg {
		pkg := &Pkg{binPath: cfg.Module.Path}
		pkg.Run()
		os.Exit(0)
	}

	// init config
	svcArgument := ""
	cfgPath := serverArgs.Cfg
	if cfgPath != "" {
		svcArgument = fmt.Sprintf("-cfg=%s", cfgPath)
	} else {
		cfgPath = filepath.Join(cfgFolder, cfgName)
	}
	_, err := os.Stat(cfgPath)
	if os.IsNotExist(err) {
		err = cfg.SaveToFile(cfgPath)
		if err != nil {
			fmt.Println("generate configure file fail: ", err)
		}
	} else {
		err = cfg.LoadFromFile(cfgPath)
		if err != nil {
			fmt.Println("load configure file fail: ", err)
		}
	}
	cfg.Path = cfgPath
	cfg.Load = cfg.DoLoad
	cfg.Save = cfg.DoSave
	cfg.InitId()

	// init certificate
	if cfg.Https.Enabled {
		certFilePath := cfg.Https.Cert.Server.File
		if certFilePath == "" {
			certFilePath = filepath.Join(rootFolder, "crt", "server.pfx")
			cfg.Https.Cert.Server.File = certFilePath
		}
	}
	if cfg.Cloud.Enabled {
		certFilePath := cfg.Cloud.Cert.Server.File
		if certFilePath == "" {
			certFilePath = filepath.Join(rootFolder, "crt", "cloud.pfx")
			cfg.Cloud.Cert.Server.File = certFilePath
		}

		certFilePath = cfg.Cloud.Cert.Ca.File
		if certFilePath == "" {
			certFilePath = filepath.Join(rootFolder, "crt", "ca.crt")
			cfg.Cloud.Cert.Ca.File = certFilePath
		}

		if cfg.Cloud.Port < 1 {
			cfg.Cloud.Port = 6931
		}
	}
	if cfg.Node.Enabled {
		certFilePath := cfg.Node.Cert.Server.File
		if certFilePath == "" {
			certFilePath = filepath.Join(rootFolder, "crt", "node.pfx")
			cfg.Node.Cert.Server.File = certFilePath
		}

		certFilePath = cfg.Node.Cert.Ca.File
		if certFilePath == "" {
			certFilePath = filepath.Join(rootFolder, "crt", "ca.crt")
			cfg.Node.Cert.Ca.File = certFilePath
		}

		if cfg.Node.CloudServer.Port < 1 {
			cfg.Node.CloudServer.Port = 6931
		}
	}

	// init path of site
	if cfg.Site.Root.Path == "" {
		cfg.Site.Root.Path = filepath.Join(rootFolder, "site", "root")
	}
	if cfg.Site.Doc.Path == "" {
		cfg.Site.Doc.Path = filepath.Join(rootFolder, "site", "doc")
	}
	if cfg.Site.Opt.Path == "" {
		cfg.Site.Opt.Path = filepath.Join(rootFolder, "site", "opt")
	}

	// init path of system service
	if cfg.Sys.Svc.Custom.App == "" {
		cfg.Sys.Svc.Custom.App = filepath.Join(rootFolder, "svc", "custom")
	}
	if cfg.Sys.Svc.Custom.Log == "" {
		cfg.Sys.Svc.Custom.Log = filepath.Join(rootFolder, "log", "svc", "custom")
	}

	// init service
	if strings.TrimSpace(cfg.Svc.Name) == "" {
		cfg.Svc.Name = moduleName
	}
	cfg.Svc.Args = svcArgument
	svcName := cfg.Svc.Name
	log.Init(cfg.Log.Level, svcName, cfg.Log.Folder)
	hdl := NewHandler(log)
	svr, err = gserver.NewServer(log, &cfg.Config, hdl)
	if err != nil {
		fmt.Println("init service fail: ", err)
		os.Exit(12)
	}
	if !svr.Interactive() {
		cfg.Svc.Restart = svr.Restart
	}
	serverArgs.Execute(svr)

	// information
	log.Std = true
	zoneName, zoneOffset := now.Zone()
	LogInfo("start at: ", moduleArgs.ModulePath())
	LogInfo("run as service: ", !svr.Interactive())
	LogInfo("version: ", moduleVersion)
	LogInfo("zone: ", zoneName, "-", zoneOffset/int(time.Hour.Seconds()))
	LogInfo("log path: ", cfg.Log.Folder)
	LogInfo("log level: ", cfg.Log.Level)
	LogInfo("configure path: ", cfgPath)
	LogInfo("configure info: ", cfg)
}
