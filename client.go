package ngorm

import (
	"errors"
	"fmt"
	"github.com/spf13/cast"
	nebula "github.com/vesoft-inc/nebula-go/v3"
	"regexp"
	"strings"
)

type Client struct {
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

func NewClient(cfg *Config) (*Client, error) {
	var (
		ok             bool
		err            error
		config         *nebula.SessionPoolConf
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

	return client, nil
}

func (c *Client) Raw(ngql string) *entity {
	var (
		e = &entity{client: c, ngql: ngql}
	)

	return e
}

func (c *Client) Fetch(ids ...string) *fetchController {
	return &fetchController{client: c, ids: ids}
}
