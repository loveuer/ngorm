package ngorm

import (
	"context"
	nebula "github.com/vesoft-inc/nebula-go/v3"
	"sync"
)

type Service struct {
	Addr string `mapstructure:"addr"`
	Port int    `mapstructure:"port"`
}

type Config struct {
	Ctx      context.Context
	Servers  []Service
	Username string
	Password string
	PoolSize int
	LogLevel LogLevel
}

type sessionChannel struct {
	idx     int
	session *nebula.Session
}

type NGDB struct {
	sync.Mutex
	ctx      context.Context
	cancel   context.CancelFunc
	space    string
	pool     *nebula.ConnectionPool
	username string
	password string
	sessChan chan *sessionChannel
}

func (db *NGDB) getSession() (*sessionChannel, error) {
	sch := <-db.sessChan
	return sch, nil
}

// Deprecated
func (db *NGDB) release(session *nebula.Session) {
	session.Release()
}

type nebulaSessionPool struct {
	sync.RWMutex
	available bool
	size      int
	ok        int
	pool      map[int]*nebula.Session
}

var (
	err  error
	pool = &nebulaSessionPool{
		ok:   0,
		pool: make(map[int]*nebula.Session),
	}
	nglog             = nebula.DefaultLogger{}
	hostList          = make([]nebula.HostAddress, 0)
	defaultPoolConfig = nebula.GetDefaultConf()
)

func (db *NGDB) initSess(size int) {
	pool.size = size
	for idx := 0; idx < size; idx++ {
		if sess, err := db.newSession(); err != nil {
			log.Warnf("init nebula session[%d] err: %v", idx, err)
			pool.pool[idx] = nil
		} else {
			log.Debugf("init nebula session[%d] success", idx)
			pool.pool[idx] = sess
			pool.ok++
			pool.available = true
		}
	}

	if pool.ok == 0 {
		pool.available = false
		log.Panic("session pool length == 0")
	}

	log.Infof("inited [%d] nebula sessions", pool.size)

	go func() {
		for {
			select {
			case <-db.ctx.Done():
				for idx := range pool.pool {
					if pool.pool[idx] != nil {
						pool.pool[idx].Release()
					}
				}
				return
			default:
				for idx := range pool.pool {
					pool.RLock()
					sess := pool.pool[idx]
					pool.RUnlock()

					db.sessChan <- &sessionChannel{idx: idx, session: sess}
				}
			}
		}
	}()
}

func (db *NGDB) newSession() (*nebula.Session, error) {
	// todo renew connection pool
	sess, err := db.pool.GetSession(db.username, db.password)
	if err == nil {
		return sess, err
	}

	log.Warnf("ngdb connection pool get session err: %v", err)

	db.Lock()
	db.pool, err = nebula.NewConnectionPool(hostList, defaultPoolConfig, nglog)
	db.Unlock()
	if err != nil {

	}

	if sess, err = db.pool.GetSession(db.username, db.password); err != nil {
		log.Errorf("ngdb connection pool get session(twice) err: %v", err)
		return sess, err
	}

	return sess, nil
}

func (db *NGDB) prepare() (*nebula.Session, error) {
	var (
		sch     *sessionChannel
		session *nebula.Session
		err     error
	)

	defer func() {
		pool.Lock()
		pool.pool[sch.idx] = session
		pool.Unlock()
	}()

	sch, _ = db.getSession()
	if sch.session == nil {
		session, err = db.newSession()
	} else {
		session = sch.session
	}

	if _, err = session.Execute("USE " + db.space); err == nil {
		return session, nil
	}

	if session, err = db.newSession(); err != nil {
		session = nil
		return session, err
	}

	if _, err = session.Execute("USE " + db.space); err != nil {
		session = nil
		return session, err
	}

	return session, nil
}

func NewNGDB(space string, config ...Config) (*NGDB, error) {
	var (
		cfg Config
		db  = new(NGDB)
	)

	if len(config) > 0 {
		cfg = config[0]
	}

	if len(cfg.Servers) == 0 {
		cfg.Servers = []Service{{Addr: "127.0.0.1", Port: 9669}}
	}

	if cfg.Username == "" {
		cfg.Username = "user"
	}

	if cfg.PoolSize == 0 {
		cfg.PoolSize = 10
	}

	if cfg.Ctx == nil {
		db.ctx, db.cancel = context.WithCancel(context.Background())
	} else {
		db.ctx, db.cancel = context.WithCancel(cfg.Ctx)
	}

	SetLogLevel(cfg.LogLevel)

	for _, s := range cfg.Servers {
		hostAddr := nebula.HostAddress{
			Host: s.Addr,
			Port: s.Port,
		}
		hostList = append(hostList, hostAddr)
	}

	defaultPoolConfig.MaxConnPoolSize = cfg.PoolSize
	defaultPoolConfig.MinConnPoolSize = cfg.PoolSize
	db.space = space
	db.username = cfg.Username
	db.password = cfg.Password

	db.pool, err = nebula.NewConnectionPool(hostList, defaultPoolConfig, nglog)
	if err != nil {
		log.Errorf("can't constructs new connection pool, err: %v", err)
		return db, err
	}

	db.sessChan = make(chan *sessionChannel)

	db.initSess(cfg.PoolSize)

	return db, nil
}

func (nd *NGDB) Close() {
	nd.cancel()
}
