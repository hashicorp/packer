package upload

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mitchellh/packer/common"
	"github.com/mitchellh/packer/helper/config"
	"github.com/mitchellh/packer/packer"
	"github.com/mitchellh/packer/template/interpolate"
	"golang.org/x/crypto/ssh"
)

type Config struct {
	common.PackerConfig `mapstructure:",squash"`
	Login               string `mapstructure:"login"`
	Passw               string `mapstructure:"passw"`
	KeyFile             string `mapstructure:"key"`
	Endpoint            string `mapstructure:"endpoint"`
	KeepInputArtifact   bool   `mapstructure:"keep_input_artifact"`
	ctx                 interpolate.Context
}

type PostProcessor struct {
	cfg Config
}

type Uploader interface {
	Upload(packer.Artifact) error
}

type UploaderSSH struct {
	Host    string
	Port    string
	Login   string
	Passw   string
	KeyFile string
	Path    string
	File    string
	Timeout time.Duration
}

func (p *PostProcessor) Configure(raws ...interface{}) error {
	err := config.Decode(&p.cfg, &config.DecodeOpts{
		Interpolate: true,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{},
		},
	}, raws...)
	if err != nil {
		return err
	}

	return nil
}

func (p *PostProcessor) PostProcess(ui packer.Ui, artifact packer.Artifact) (packer.Artifact, bool, error) {
	var err error
	var upl Uploader

	newartifact := &Artifact{endpoint: p.cfg.Endpoint}
	//ssh://vtolstov:xxx@rl01.01.mighost.ru:2022/srv/vps/?key=/home/vtolstov/.ssh/mighost&timeout=20s
	u, err := url.Parse(p.cfg.Endpoint)
	if err != nil {
		return nil, p.cfg.KeepInputArtifact, err
	}

	switch u.Scheme {
	case "ssh":
		upl, err = p.newUploaderSSH(u)
	}
	if err != nil {
		return nil, p.cfg.KeepInputArtifact, err
	}

	err = upl.Upload(artifact)

	return newartifact, p.cfg.KeepInputArtifact, err
}

func (p *PostProcessor) newUploaderSSH(u *url.URL) (*UploaderSSH, error) {
	var err error
	upl := &UploaderSSH{}

	if u.User != nil {
		if pwd, ok := u.User.Password(); ok {
			upl.Passw = pwd
		}
		upl.Login = u.User.Username()
	}
	if strings.Index(u.Host, ":") > 0 {
		upl.Host, upl.Port, err = net.SplitHostPort(u.Host)
		if err != nil {
			return nil, err
		}
	} else {
		upl.Host = u.Host
		upl.Port = "22"
	}

	if strings.HasSuffix(u.Path, "/") {
		upl.Path = u.Path
	} else {
		upl.Path = filepath.Dir(u.Path)
		upl.File = filepath.Base(u.Path)
	}
	if u.RawQuery != "" {
		parts := strings.Split(u.RawQuery, "&")
		for _, part := range parts {
			p := strings.Split(part, "=")
			switch p[0] {
			case "timeout":
				upl.Timeout, err = time.ParseDuration(p[1])
				if err != nil {
					return nil, err
				}
			case "keyfile":
				upl.KeyFile = p[1]
			}
		}
	}
	return upl, nil
}

func (u *UploaderSSH) Upload(artifact packer.Artifact) error {
	if len(artifact.Files()) > 1 && u.File != "" {
		return errors.New("endpoint path is file, but artifact contains multiple files")
	}
	clientConfig := &ssh.ClientConfig{
		User: u.Login,
		Auth: []ssh.AuthMethod{},
	}
	if u.KeyFile != "" {
		privateKey, err := ioutil.ReadFile(u.KeyFile)
		if err != nil {
			return err
		}
		signer, _ := ssh.ParsePrivateKey([]byte(privateKey))
		clientConfig.Auth = append(clientConfig.Auth, ssh.PublicKeys(signer))
	}
	if u.Passw != "" {
		clientConfig.Auth = append(clientConfig.Auth, ssh.Password(u.Passw))
	}
	client, err := ssh.Dial("tcp", u.Host+":"+u.Port, clientConfig)
	if err != nil {
		return err
	}
	defer client.Close()
	session, err := client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()
	go func() error {
		w, _ := session.StdinPipe()
		defer w.Close()
		for _, art := range artifact.Files() {
			fp, err := os.Open(art)
			if err != nil {
				return err
			}
			fi, err := fp.Stat()
			if err != nil {
				return err
			}
			fmt.Fprintln(w, "D0755", 0, u.Path)
			if u.File != "" {
				fmt.Fprintln(w, "C0644", fi.Size(), u.File)
			} else {
				fmt.Fprintln(w, "C0644", fi.Size(), fi.Name())
			}
			_, err = io.Copy(w, fp)
			if err != nil {
				return err
			}
			fmt.Fprint(w, "\x00")
		}
		return nil
	}()
	err = session.Run("/usr/bin/scp -tr ./")
	return err
}
