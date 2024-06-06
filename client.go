package ngorm

import (
	"context"
	"errors"
	"fmt"
	"github.com/loveuer/esgo2dump/log"
	"github.com/spf13/cast"
	nebula "github.com/vesoft-inc/nebula-go/v3"
	"regexp"
	"strings"
)

type Client struct {
	ctx    context.Context
	pool   *nebula.SessionPool
	logger logger
}

type Config struct {
	Endpoints    []string
	Username     string
	Password     string
	DefaultSpace string
	Logger       logger
}

var (
	config    *nebula.SessionPoolConf
	cc        *Config
	domainReg *regexp.Regexp
	ipv4Reg   *regexp.Regexp
)

func init() {
	var err error
	domainReg, err = regexp.Compile(`^(([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9])\.)*([A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\-]*[A-Za-z0-9])$`)
	if err != nil {
		log.Panic(err.Error())
	}

	ipv4Reg, err = regexp.Compile(`^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])$`)
	if err != nil {
		log.Panic(err.Error())
	}
}

func NewClient(ctx context.Context, cfg *Config) (*Client, error) {
	var (
		//ok             bool
		err            error
		pool           *nebula.SessionPool
		serviceAddress = make([]nebula.HostAddress, 0)
		client         = &Client{}
	)

	if cfg == nil {
		return nil, errors.New("config is nil")
	}

	if cfg.Logger != nil {
		clog.l = cfg.Logger
	}

	if ctx == nil {
		ctx = context.Background()
	}

	for _, endpoint := range cfg.Endpoints {
		s := strings.Split(endpoint, ":")
		if len(s) != 2 {
			clog.l.Warn("endpoint: %s invalid", endpoint)
			continue
		}

		var (
			host = s[0]
			port int
		)

		if port = cast.ToInt(s[1]); port == 0 {
			clog.l.Warn("endpoint: %s invalid", endpoint)
			continue
		}

		if (!ipv4Reg.MatchString(host)) && (!domainReg.MatchString(host)) {
			clog.l.Warn("endpoint: %s invalid", endpoint)
			continue
		}

		clog.l.Debug("add endpoint: host: %s, port: %d", host, port)
		serviceAddress = append(serviceAddress, nebula.HostAddress{Host: host, Port: port})
	}

	if config, err = nebula.NewSessionPoolConf(
		cfg.Username,
		cfg.Password,
		serviceAddress,
		cfg.DefaultSpace,
		nebula.WithMaxSize(50),
		nebula.WithMinSize(10),
	); err != nil {
		cfg.Logger.Debug(fmt.Sprintf("[ngorm] new session pool conf err: %v", err))
		return nil, err
	}

	if pool, err = nebula.NewSessionPool(*config, clog); err != nil {
		clog.l.Debug("new session pool err: %v", err)
		return nil, fmt.Errorf("new session pool err: %w", err)
	}

	client.pool = pool
	client.logger = cfg.Logger
	client.ctx = ctx

	cc = cfg

	return client, nil
}

func (c *Client) Session(scfgs ...*SessCfg) *Session {
	if len(scfgs) > 0 && scfgs[0] != nil {
		return &Session{client: c, cfg: scfgs[0]}
	}

	return &Session{client: c, cfg: sessionDefaultCfg}
}

// Deprecated: use Session().Raw instead
func (c *Client) Raw(ngql string) *entity {
	sess := &Session{client: c, cfg: sessionDefaultCfg}
	e := &entity{sess: sess, ngql: ngql}

	return e
}

// Deprecated: use Session().Fetch instead
func (c *Client) Fetch(ids ...string) *fetchController {
	sess := &Session{client: c, cfg: sessionDefaultCfg}
	return &fetchController{sess: sess, ids: ids}
}

// Deprecated: use Session().GoFrom instead
func (c *Client) GoFrom(id string) *goController {
	sess := &Session{client: c, cfg: sessionDefaultCfg}
	return &goController{sess: sess, from: id}
}

// Deprecated: use Session().Match instead
func (c *Client) Match(value any, name string) *matchController {
	sess := &Session{client: c, cfg: sessionDefaultCfg}
	return &matchController{
		sess:   sess,
		points: []drop{{Name: name, Value: value}},
		edgs:   make(map[string]struct{}),
	}
}
