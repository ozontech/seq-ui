package tokenlimiter

import "sync"

type Limiter struct {
	maxTokens int

	m  map[string]chan struct{}
	mu *sync.Mutex
}

func New(maxTokens int) *Limiter {
	return &Limiter{
		maxTokens: maxTokens,
		m:         make(map[string]chan struct{}, 0),
		mu:        new(sync.Mutex),
	}
}

func (l *Limiter) Limited(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	if _, ok := l.m[key]; !ok {
		l.m[key] = make(chan struct{}, l.maxTokens)
		for i := 0; i < l.maxTokens; i++ {
			l.m[key] <- struct{}{}
		}
	}

	select {
	case <-l.m[key]:
		return false
	default:
		return true
	}
}

func (l *Limiter) Fill(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if c, ok := l.m[key]; !ok || len(c) == l.maxTokens {
		return
	}

	l.m[key] <- struct{}{}
}
