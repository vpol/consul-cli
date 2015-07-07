package command

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"

	consulapi "github.com/hashicorp/consul/api"
	flag "github.com/ogier/pflag"
	"github.com/mitchellh/cli"
)

type Meta struct {
	UI		cli.Ui
	consulAddr	string
	sslEnabled	bool
	sslVerify	bool
	sslCert		string
	sslCaCert	string
	token		string
	auth		*Auth
	dc		string
}

func (m *Meta) FlagSet() *flag.FlagSet {
	f := flag.NewFlagSet("consul-cli", flag.ContinueOnError)
	f.StringVar(&m.consulAddr, "consul", "127.0.0.1:8500", "")
	f.BoolVar(&m.sslEnabled, "ssl", false, "")
	f.BoolVar(&m.sslVerify, "ssl-verify", true, "")
	f.StringVar(&m.sslCert, "ssl-cert", "", "")
	f.StringVar(&m.sslCaCert, "ssl-ca-cert", "", "")
	f.StringVar(&m.token, "token", "", "")
	f.StringVar(&m.dc, "datacenter", "", "")

	m.auth = new(Auth)
	f.Var((*Auth)(m.auth), "auth", "")

	return f
}

func (m *Meta) Client() (*consulapi.Client, error) {
	config := consulapi.DefaultConfig()
	config.Address = m.consulAddr

	if m.token != "" {
		config.Token = m.token
	}

	if m.sslEnabled {
		config.Scheme = "https"
	}

	if !m.sslVerify {
		config.HttpClient.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}
	}

	if m.auth.Enabled {
		config.HttpAuth = &consulapi.HttpBasicAuth{
			Username: m.auth.Username,
			Password: m.auth.Password,
		}
	}

	client, err := consulapi.NewClient(config)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func (m *Meta) WriteOptions() *consulapi.WriteOptions {
	writeOpts := new(consulapi.WriteOptions)
	if m.token != "" {
		writeOpts.Token = m.token
	}

	if m.dc != "" {
		writeOpts.Datacenter = m.dc
	}

	return writeOpts
}

func (m *Meta) QueryOptions() *consulapi.QueryOptions {
	queryOpts := new(consulapi.QueryOptions)
	if m.token != "" {
		queryOpts.Token = m.token
	}

	if m.dc != "" {
		queryOpts.Datacenter = m.dc
	}

	return queryOpts
}

func (m *Meta) ConsulHelp() string {
	helpText := `
  --consul=127.0.0.1:8500	HTTP address of the Consul Agent
  --ssl				Use HTTPS when talking to Consul
				(default: false)
  --ssl-verify			Verify certificates when connecting via SSL
				(default: true)
  --ssl-cert			Path to an SSL certificate to use to authenticate
				to the Consul server
				(default: not set)
  --ssl-ca-cert			Path to an SSL client certificate to use to authenticate
				to the Consul server
				(default: not set)
  --token			The Consul ACL token
				(default: not set)
  --datacenter			Consul data center
				(default: not set)
`

  return helpText
}

// Authentication var
type Auth struct {
	Enabled		bool
	Username	string
	Password	string
}


func (a *Auth) Set(value string) error {
	a.Enabled = true

	if (strings.Contains(value, ":")) {
		split := strings.SplitN(value, ":", 2)
		a.Username = split[0]
		a.Password = split[1]
	} else {
		a.Username = value
	}

	return nil
}

func (a *Auth) String() string {
	if a.Password == "" {
		return a.Username
	}

	return fmt.Sprintf("%s:%s", a.Username, a.Password)
}

type funcVar func(s string) error

func (f funcVar) Set(s string) error	{ return f(s) }
func (f funcVar) String() string	{ return "" }
func (f funcVar) IsBoolFlag() bool	{ return false }