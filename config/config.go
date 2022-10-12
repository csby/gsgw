package config

import (
	"encoding/json"
	"fmt"
	"github.com/csby/gwsf/gcfg"
	"github.com/csby/gwsf/gtype"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
)

type Config struct {
	sync.RWMutex
	gcfg.Config
}

func NewConfig() *Config {
	return &Config{
		Config: gcfg.Config{
			Log: gcfg.Log{
				Folder: "",
				Level:  "error|warning|info",
			},
			Node: gcfg.Node{
				Enabled: false,
				CloudServer: gcfg.Cloud{
					Address: "",
					Port:    6931,
				},
				Cert: gcfg.Crt{
					Ca: gcfg.CrtCa{
						File: "",
					},
					Server: gcfg.CrtPfx{
						File:     "",
						Password: "",
					},
				},
				Forward: gcfg.NodeFwd{
					Enable: false,
					Items:  []*gcfg.Fwd{},
				},
			},
			Http: gcfg.Http{
				Enabled:     true,
				Port:        8080,
				BehindProxy: false,
			},
			Https: gcfg.Https{
				Enabled:     false,
				Port:        8443,
				BehindProxy: false,
				Cert: gcfg.Crt{
					Ca: gcfg.CrtCa{
						File: "",
					},
					Server: gcfg.CrtPfx{
						File:     "",
						Password: "",
					},
				},
				RequestClientCert: false,
			},
			Cloud: gcfg.Https{
				Enabled:     false,
				Port:        6931,
				BehindProxy: false,
				Cert: gcfg.Crt{
					Ca: gcfg.CrtCa{
						File: "",
					},
					Server: gcfg.CrtPfx{
						File:     "",
						Password: "",
					},
				},
				RequestClientCert: true,
			},
			Site: gcfg.Site{
				Doc: gcfg.SiteDoc{
					Enabled:       true,
					DownloadTitle: "从github下载",
					DownloadUrl:   "https://github.com/csby/gwsf-doc/releases",
				},
				Opt: gcfg.SiteOpt{
					Users: []*gcfg.SiteOptUser{
						{
							Account:  "admin",
							Password: "admin",
							Name:     "管理员",
						},
					},
					Ldap: gcfg.SiteOptLdap{
						Host: "127.0.0.1",
						Port: 389,
						Base: "dc=dev,dc=local",
					},
				},
			},
			Proxy: "",
			ReverseProxy: gcfg.Proxy{
				Enabled: false,
				Servers: []*gcfg.ProxyServer{
					{
						Id:      gtype.NewGuid(),
						Name:    "http",
						Disable: true,
						TLS:     false,
						IP:      "",
						Port:    "80",
						Targets: []*gcfg.ProxyTarget{},
					},
					{
						Id:      gtype.NewGuid(),
						Name:    "https",
						Disable: true,
						TLS:     true,
						IP:      "",
						Port:    "443",
						Targets: []*gcfg.ProxyTarget{},
					},
				},
			},
			Sys: gcfg.System{
				Svc: gcfg.Service{
					Enabled: false,
					Custom: gcfg.ServiceCustom{
						DownloadTitle: "从github下载",
						DownloadUrl:   "https://github.com/csby/gtool/releases",
					},
					Tomcats: []*gcfg.ServiceTomcat{},
					Others:  []*gcfg.ServiceOther{},
					Nginxes: []*gcfg.ServiceNginx{},
					Files:   []*gcfg.ServiceFile{},
				},
			},
		},
	}
}

func (s *Config) LoadFromFile(filePath string) error {
	s.Lock()
	defer s.Unlock()

	bytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	return json.Unmarshal(bytes, s)
}

func (s *Config) SaveToFile(filePath string) error {
	s.Lock()
	defer s.Unlock()

	bytes, err := json.MarshalIndent(s, "", "    ")
	if err != nil {
		return err
	}

	fileFolder := filepath.Dir(filePath)
	_, err = os.Stat(fileFolder)
	if os.IsNotExist(err) {
		os.MkdirAll(fileFolder, 0777)
	}

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = fmt.Fprint(file, string(bytes[:]))

	return err
}

func (s *Config) DoLoad() (*gcfg.Config, error) {
	c := &Config{}
	e := c.LoadFromFile(s.Path)
	if e != nil {
		return nil, e
	}

	return &c.Config, nil
}

func (s *Config) DoSave(cfg *gcfg.Config) error {
	if cfg == nil {
		return nil
	}

	c := &Config{}
	e := c.LoadFromFile(s.Path)
	if e != nil {
		return e
	}

	c.Config = *cfg
	return c.SaveToFile(s.Path)
}

func (s *Config) String() string {
	bytes, err := json.Marshal(s)
	if err != nil {
		return ""
	}

	return string(bytes[:])
}

func (s *Config) FormatString() string {
	bytes, err := json.MarshalIndent(s, "", "    ")
	if err != nil {
		return ""
	}

	return string(bytes[:])
}
