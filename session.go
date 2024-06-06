package ngorm

type SessCfg struct {
	Debug    bool
	Logger   logger
	MaxRetry int
}

type Session struct {
	cfg    *SessCfg
	client *Client
	logger logger
}

var (
	sessionDefaultCfg = &SessCfg{
		Debug:    false,
		Logger:   clog.l,
		MaxRetry: 1,
	}
)

func (s *Session) Debug() *Session {
	s.cfg.Debug = true

	return s
}

func (s *Session) Raw(ngql string) *entity {
	e := &entity{sess: s, ngql: ngql}
	return e
}

func (s *Session) Fetch(ids ...string) *fetchController {
	return &fetchController{
		sess: s,
		ids:  ids,
	}
}

func (s *Session) GoFrom(id string) *goController {
	return &goController{
		sess: s,
		from: id,
	}
}

func (s *Session) Match(value any, name string) *matchController {
	return &matchController{
		sess:   s,
		points: []drop{{Name: name, Value: value}},
		edgs:   make(map[string]struct{}),
	}
}
