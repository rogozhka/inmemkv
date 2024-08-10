package inmemkv

import (
    "sync"
    "sync/atomic"
    "time"
)

type inMemKeyValue struct {
    mm          *sync.Map // key => value
    mmDeadlines *sync.Map // key => deadline

    isExpirable atomic.Bool
    ttl         atomic.Value
}

func NewCache(opts ...KeyValueOpt) KeyValueStorage {
    p := &inMemKeyValue{}
    for _, opt := range opts {
        opt(p)
    }
    p.Reset()
    return p
}

type KeyValueOpt func(p *inMemKeyValue)

func WithTTL(ttl time.Duration) KeyValueOpt {
    return func(p *inMemKeyValue) {
        p.ttl.Store(ttl)
        p.isExpirable.Store(true)
    }
}

func (p *inMemKeyValue) Get(key string) (any, bool) {
    v, ok := p.mm.Load(key)
    if !ok {
        return nil, false
    }
    if p.isExpirable.Load() {
        raw, ok := p.mmDeadlines.Load(key)
        if !ok {
            p.cleanup(key)
            return nil, false
        }
        d := raw.(time.Time)
        if time.Now().After(d) {
            p.cleanup(key)
            return nil, false
        }
    }
    return v, ok
}

func (p *inMemKeyValue) cleanup(key string) {
    p.mm.Delete(key)
    if p.isExpirable.Load() {
        p.mmDeadlines.Delete(key)
    }
}

func (p *inMemKeyValue) Set(key string, value any) {
    p.mm.Store(key, value)
    if p.isExpirable.Load() {
        ttl := p.ttl.Load().(time.Duration)
        p.mmDeadlines.Store(key, time.Now().Add(ttl))
    }
}

func (p *inMemKeyValue) Delete(key string) {
    p.cleanup(key)
}

func (p *inMemKeyValue) Is(key string) bool {
    _, ok := p.Get(key)
    return ok
}

func (p *inMemKeyValue) Reset() {
    p.mm = &sync.Map{}
    if p.isExpirable.Load() {
        p.mmDeadlines = &sync.Map{}
    }
}
