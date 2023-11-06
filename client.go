package ngorm

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/spf13/cast"
	nebula "github.com/vesoft-inc/nebula-go/v3"
)

type Client struct {
	ctx    context.Context
	client *nebula.SessionPool
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
	config *nebula.SessionPoolConf
	cc     *Config
)

func NewClient(ctx context.Context, cfg *Config) (*Client, error) {
	var (
		ok             bool
		err            error
		pool           *nebula.SessionPool
		serviceAddress = make([]nebula.HostAddress, 0)
		client         = &Client{}
	)

	if cfg == nil {
		DefaultLogger.Debug("[ngorm] nil config")
		return nil, errors.New("config is nil")
	}

	if cfg.Logger == nil {
		cfg.Logger = DefaultLogger
	}

	if ctx == nil {
		ctx = context.Background()
	}

	for _, endpoint := range cfg.Endpoints {
		s := strings.Split(endpoint, ":")
		if len(s) != 2 {
			cfg.Logger.Debug(fmt.Sprintf("[ngorm] endpoint: %s invalid", endpoint))
			continue
		}

		var (
			host = s[0]
			port int
		)

		if port = cast.ToInt(s[1]); port == 0 {
			cfg.Logger.Debug(fmt.Sprintf("[ngorm] endpoint: %s invalid", endpoint))
			continue
		}

		if ok, _ = regexp.MatchString(`^((25[0-5]|(2[0-4]|1\d|[1-9]|)\d)\.?\b){4}$`, host); !ok {
			cfg.Logger.Debug(fmt.Sprintf("[ngorm] endpoint: %s invalid", endpoint))
			continue
		}

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

	if pool, err = nebula.NewSessionPool(*config, cfg.Logger); err != nil {
		cfg.Logger.Debug(fmt.Sprintf("[ngorm] new session pool err: %v", err))
		return nil, err
	}

	client.client = pool
	client.logger = cfg.Logger
	client.ctx = ctx

	cc = cfg

	return client, nil
}

func (c *Client) Raw(ngql string) *entity {
	var (
		e = &entity{c: c, ngql: ngql, logger: c.logger}
	)

	return e
}

func (c *Client) Fetch(ids ...string) *fetchController {
	return &fetchController{client: c, ids: ids}
}

func (c *Client) GoFrom(id string) *goController {
	return &goController{client: c, from: id}
}

func (c *Client) MatchHead(value any) *matchController {
	return &matchController{client: c, head: drop{Number: "head", Value: value}}
}
