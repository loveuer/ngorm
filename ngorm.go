package ngorm

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	nebula "github.com/vesoft-inc/nebula-go/v3"
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

type NGDB struct {
	sync.Mutex
	ctx      context.Context
	cancel   context.CancelFunc
	space    string
	pool     *nebula.ConnectionPool
	username string
	password string
	timeout  int
}

func (db *NGDB) getSession(ctx context.Context) (*nebula.Session, error) {
	var (
		err     error
		ch      = make(chan *nebula.Session)
		fctx, _ = context.WithTimeout(ctx, 30*time.Second)
	)

	go func() {
		var (
			sess *nebula.Session
		)

		for {
			if sess, err = db.newSession(); err != nil {
				logrus.Debugf("nubela new session err: %v", err)
				continue
			}

			break
		}

		ch <- sess
	}()

	select {
	case newSess := <-ch:
		return newSess, nil
	case <-fctx.Done():
		return nil, errors.New("acquire new nebula session timeout")
	}
}

func (db *NGDB) newSession() (*nebula.Session, error) {
	return db.pool.GetSession(db.username, db.password)
}

func (db *NGDB) prepare(ctx context.Context) (*nebula.Session, error) {
	var (
		err  error
		sess *nebula.Session
	)

	if sess, err = db.getSession(ctx); err != nil {
		return nil, err
	}

	if _, err = sess.Execute("USE " + db.space); err != nil {
		return nil, err
	}

	return sess, nil
}

func NewNGDB(space string, config ...Config) (*NGDB, error) {
	var (
		err      error
		cfg      Config
		db       = new(NGDB)
		hostList = make([]nebula.HostAddress, 0)
		//nglog             = logrus.New()
		nglog             = nebula.DefaultLogger{}
		defaultPoolConfig = nebula.GetDefaultConf()
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

	log.Infof("inited nebula connection pool(size: %d)", defaultPoolConfig.MinConnPoolSize)

	return db, nil
}

func (nd *NGDB) Close() {
	nd.cancel()
}
