package cache

import (
	"bytes"
	"crypto/sha1"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/sync/singleflight"

	"github.com/things-go/gin-cache/persist"
)

// PageCachePrefix default page cache key prefix
var PageCachePrefix = "gincache.page.cache"

// Logger logger interface
type Logger interface {
	Errorf(format string, args ...interface{})
}

// Pool BodyCache pool
type Pool interface {
	Get() *BodyCache
	Put(*BodyCache)
}

// Config config for cache
type Config struct {
	// store the cache backend to store response
	store persist.Store
	// expire the cache expiration time
	expire time.Duration
	// rand rand duration for expire
	rand func() time.Duration
	// generate key for store, bool means need cache or not
	generateKey func(c *gin.Context) (string, bool)
	// group single flight group
	group *singleflight.Group
	// BodyCache pool
	pool Pool
	// logger debug
	logger Logger
}

// Option custom option
type Option func(c *Config)

// WithGenerateKey custom generate key ,default is GenerateRequestURIKey.
func WithGenerateKey(f func(c *gin.Context) (string, bool)) Option {
	return func(c *Config) {
		if f != nil {
			c.generateKey = f
		}
	}
}

// WithSingleflight custom single flight group, default is private single flight group.
func WithSingleflight(group *singleflight.Group) Option {
	return func(c *Config) {
		if group != nil {
			c.group = group
		}
	}
}

// WithBodyCachePool custom body cache pool, default is private cache pool.
func WithBodyCachePool(p Pool) Option {
	return func(c *Config) {
		if p != nil {
			c.pool = p
		}
	}
}

// WithRandDuration custom rand duration for expire, default return zero
// expiration time always expire + rand()
func WithRandDuration(rand func() time.Duration) Option {
	return func(c *Config) {
		if rand != nil {
			c.rand = rand
		}
	}
}

// WithLogger custom logger, default is Discard.
func WithLogger(l Logger) Option {
	return func(c *Config) {
		if l != nil {
			c.logger = l
		}
	}
}

// Cache user must pass store and store expiration time to cache and with custom option.
// default caching response with uri, which use PageCachePrefix
func Cache(store persist.Store, expire time.Duration, handle gin.HandlerFunc, opts ...Option) gin.HandlerFunc {
	cfg := Config{
		store:       store,
		expire:      expire,
		rand:        func() time.Duration { return 0 },
		generateKey: GenerateRequestURIKey,
		group:       new(singleflight.Group),
		pool:        NewPool(),
		logger:      NewDiscard(),
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	return func(c *gin.Context) {
		key, needCache := cfg.generateKey(c)
		if !needCache {
			handle(c)
			return
		}

		// read cache first
		bodyCache := cfg.pool.Get()
		defer cfg.pool.Put(bodyCache)

		if err := cfg.store.Get(key, &bodyCache); err != nil {
			// BodyWriter in order to dup the response
			bodyWriter := &BodyWriter{ResponseWriter: c.Writer}
			c.Writer = bodyWriter

			inFlight := false
			// use single flight to avoid Hotspot Invalid
			bc, _, shared := cfg.group.Do(key, func() (interface{}, error) {
				handle(c)
				inFlight = true
				bc := getBodyCacheFromBodyWriter(bodyWriter)
				if !c.IsAborted() && bodyWriter.Status() < 300 && bodyWriter.Status() >= 200 {
					if err = cfg.store.Set(key, bc, cfg.expire+cfg.rand()); err != nil {
						cfg.logger.Errorf("set cache key error: %s, cache key: %s", err, key)
					}
				}
				return bc, nil
			})
			if !inFlight && shared {
				responseWithBodyCache(c, bc.(*BodyCache))
			}
		} else {
			responseWithBodyCache(c, bodyCache)
		}
	}
}

// CacheWithRequestURI a shortcut function for caching response with uri
func CacheWithRequestURI(store persist.Store, expire time.Duration, handle gin.HandlerFunc, opts ...Option) gin.HandlerFunc {
	return Cache(store, expire, handle, opts...)
}

// CacheWithRequestPath a shortcut function for caching response with url path, which discard the query params
func CacheWithRequestPath(store persist.Store, expire time.Duration, handle gin.HandlerFunc, opts ...Option) gin.HandlerFunc {
	return Cache(store, expire, handle, append(opts, WithGenerateKey(GenerateRequestPathKey))...)
}

// GenerateKeyWithPrefix generate key with GenerateKeyWithPrefix and u,
// if key is larger than 200,it will use sha1.Sum
// key like: prefix:u or prefix:sha1(u)
func GenerateKeyWithPrefix(prefix, key string) string {
	if len(key) > 200 {
		d := sha1.Sum([]byte(key))
		return prefix + ":" + string(d[:])
	}
	return prefix + ":" + key
}

// GenerateRequestURIKey generate key with PageCachePrefix and request uri
func GenerateRequestURIKey(c *gin.Context) (string, bool) {
	return GenerateKeyWithPrefix(PageCachePrefix, url.QueryEscape(c.Request.RequestURI)), true
}

// GenerateRequestPathKey generate key with PageCachePrefix and request Path
func GenerateRequestPathKey(c *gin.Context) (string, bool) {
	return GenerateKeyWithPrefix(PageCachePrefix, url.QueryEscape(c.Request.URL.Path)), true
}

// BodyWriter dup response writer body
type BodyWriter struct {
	gin.ResponseWriter
	dupBody bytes.Buffer
}

// Write writes the data to the connection as part of an HTTP reply.
func (w *BodyWriter) Write(b []byte) (int, error) {
	w.dupBody.Write(b)
	return w.ResponseWriter.Write(b)
}

// WriteString the string into the response body.
func (w *BodyWriter) WriteString(s string) (int, error) {
	w.dupBody.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}

// BodyCache body cache store
type BodyCache struct {
	Status int
	Header http.Header
	Data   []byte
}

func getBodyCacheFromBodyWriter(writer *BodyWriter) *BodyCache {
	return &BodyCache{
		writer.Status(),
		writer.Header().Clone(),
		writer.dupBody.Bytes(),
	}
}

func responseWithBodyCache(c *gin.Context, bodyCache *BodyCache) {
	c.Writer.WriteHeader(bodyCache.Status)
	for k, v := range bodyCache.Header {
		for _, vv := range v {
			c.Writer.Header().Add(k, vv)
		}
	}
	c.Writer.Write(bodyCache.Data) // nolint: errcheck
}

type cachePool struct {
	pool *sync.Pool
}

// NewPool new pool for BodyCache
func NewPool() Pool {
	return &cachePool{
		&sync.Pool{
			New: func() interface{} { return &BodyCache{Header: make(http.Header)} },
		},
	}
}

// Get implement Pool interface
func (sf *cachePool) Get() *BodyCache {
	return sf.pool.Get().(*BodyCache)
}

// Put implement Pool interface
func (sf *cachePool) Put(c *BodyCache) {
	c.Data = c.Data[:0]
	c.Header = make(http.Header)
	sf.pool.Put(c)
}

// Discard is an logger on which all Write calls succeed
// without doing anything.
type Discard struct{}

var _ Logger = (*Discard)(nil)

// NewDiscard a discard logger on which always succeed without doing anything
func NewDiscard() Discard { return Discard{} }

// Debugf implement Logger interface.
func (sf Discard) Debugf(string, ...interface{}) {}

// Infof implement Logger interface.
func (sf Discard) Infof(string, ...interface{}) {}

// Errorf implement Logger interface.
func (sf Discard) Errorf(string, ...interface{}) {}

// Warnf implement Logger interface.
func (sf Discard) Warnf(string, ...interface{}) {}

// DPanicf implement Logger interface.
func (sf Discard) DPanicf(string, ...interface{}) {}

// Fatalf implement Logger interface.
func (sf Discard) Fatalf(string, ...interface{}) {}
