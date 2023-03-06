package util

import (
	"crypto/tls"
	"net/http"
	"os"
	"strings"

	"github.com/containerd/containerd/remotes"
	"github.com/containerd/containerd/remotes/docker"
	"github.com/docker/cli/cli/config"
	"github.com/docker/cli/cli/config/configfile"
	"github.com/docker/cli/cli/config/credentials"
	"github.com/sirupsen/logrus"
)

func NewResolver(username, password string, insecure, plainHTTP bool, dockerConfigPath string) remotes.Resolver {

	opts := docker.ResolverOptions{
		PlainHTTP: plainHTTP,
	}
	client := http.DefaultClient
	if insecure {
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}
	}
	opts.Client = client

	if username != "" || password != "" {
		opts.Credentials = func(hostName string) (string, string, error) {
			return username, password, nil
		}
		return docker.NewResolver(opts)
	}
	var (
		err error
		cfg *configfile.ConfigFile
	)
	if dockerConfigPath == "" || dockerConfigPath == config.Dir() {
		cfg, err = config.Load(config.Dir())
		if err != nil {
			// handle error
			logrus.Errorf("unable to load default Docker auth config: %v", err)
		}
	} else {
		cfg = configfile.New(dockerConfigPath)
		if _, err := os.Stat(dockerConfigPath); err == nil {
			file, err := os.Open(dockerConfigPath)
			if err != nil {
				logrus.Errorf("Can't load docker config file %s: %v", dockerConfigPath, err)
				// fall back to resolver with no config
				return docker.NewResolver(opts)
			}
			defer file.Close()
			if err := cfg.LoadFromReader(file); err != nil {
				logrus.Errorf("Can't read and parse docker config file %s: %v", dockerConfigPath, err)
				return docker.NewResolver(opts)
			}
		} else if !os.IsNotExist(err) {
			logrus.Errorf("Unable to open docker config file %s: %v", dockerConfigPath, err)
			return docker.NewResolver(opts)
		}
	}
	if !cfg.ContainsAuth() {
		cfg.CredentialsStore = credentials.DetectDefaultStore(cfg.CredentialsStore)
	}
	// support cred helpers
	opts.Credentials = func(hostName string) (string, string, error) {
		hostname := resolveHostname(hostName)
		auth, err := cfg.GetAuthConfig(hostname)
		if err != nil {
			return "", "", err
		}
		if auth.IdentityToken != "" {
			return "", auth.IdentityToken, nil
		}
		return auth.Username, auth.Password, nil
	}
	return docker.NewResolver(opts)
}

// resolveHostname resolves Docker specific hostnames
func resolveHostname(hostname string) string {
	if strings.HasSuffix(hostname, "docker.io") {
		// Docker's `config.json` uses index.docker.io as the reference
		return LegacyDefaultHostname
	}
	return hostname
}
