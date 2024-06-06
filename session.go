package ngorm

type SessCfg struct {
	Debug    bool
	Logger   logger
	MaxRetry int
}

type session struct {
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

func (s *session) Debug() *session {
	s.cfg.Debug = true

	return s
}

func (s *session) Raw(ngql string) *entity {
	e := &entity{sess: s, ngql: ngql}
	return e
}

func (s *session) Fetch(ids ...string) *fetchController {
	return &fetchController{
		sess: s,
		ids:  ids,
	}
}

func (s *session) GoFrom(id string) *goController {
	return &goController{
		sess: s,
		from: id,
	}
}

func (s *session) Match(value any, name string) *matchController {
	return &matchController{
		sess:   s,
		points: []drop{{Name: name, Value: value}},
		edgs:   make(map[string]struct{}),
	}
}
