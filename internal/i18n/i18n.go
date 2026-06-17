package i18n

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"path"
)

type Locale struct {
	Lang   string
	labels map[string]string
}

func (l *Locale) T(key string, args ...interface{}) string {
	s, ok := l.labels[key]
	if !ok {
		return key
	}
	if len(args) > 0 {
		return fmt.Sprintf(s, args...)
	}
	return s
}

//go:embed locales/*.json
var localeFS embed.FS

type Store struct {
	locales     map[string]*Locale
	defaultLang string
}

func NewStore() *Store {
	return &Store{
		locales:     make(map[string]*Locale),
		defaultLang: "pt-BR",
	}
}

func (s *Store) Load() error {
	entries, err := localeFS.ReadDir("locales")
	if err != nil {
		return fmt.Errorf("reading locales dir: %w", err)
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		data, err := localeFS.ReadFile(path.Join("locales", e.Name()))
		if err != nil {
			return fmt.Errorf("reading locale file %s: %w", e.Name(), err)
		}
		var labels map[string]string
		if err := json.Unmarshal(data, &labels); err != nil {
			return fmt.Errorf("parsing locale file %s: %w", e.Name(), err)
		}
		lang := e.Name()[:len(e.Name())-5]
		s.locales[lang] = &Locale{Lang: lang, labels: labels}
	}
	return nil
}

func (s *Store) Get(lang string) *Locale {
	if l, ok := s.locales[lang]; ok {
		return l
	}
	return s.locales[s.defaultLang]
}

func (s *Store) Default() *Locale {
	return s.locales[s.defaultLang]
}

type ctxKey struct{}

func ToContext(ctx context.Context, loc *Locale) context.Context {
	return context.WithValue(ctx, ctxKey{}, loc)
}

func FromContext(ctx context.Context) *Locale {
	loc, _ := ctx.Value(ctxKey{}).(*Locale)
	return loc
}

func T(ctx context.Context, key string, args ...interface{}) string {
	loc := FromContext(ctx)
	if loc == nil {
		return key
	}
	return loc.T(key, args...)
}
