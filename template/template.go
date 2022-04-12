package template

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/pkg/errors"
)

var (
	defaultTemplates = &Templates{
		Cache:  make(map[string]*Template),
		Lock:   &sync.RWMutex{},
		Update: make(chan *Template, 20),
	}
)

const (
	TPL_TYPE_FILE = iota
	TPL_TYPE_STRING
)

type Config struct {
}

type Templates struct {
	Cache  map[string]*Template
	Lock   *sync.RWMutex
	Update chan *Template
}

func (ts *Templates) Watch() {
	timer := time.NewTicker(time.Millisecond * 1000)
	for {
		select {
		case <-timer.C:
			for _, t := range ts.Cache {
				go ts.checkVersion(t)
			}
		case t := <-ts.Update:
			t.update()
		}
	}
}

func (ts *Templates) checkVersion(t *Template) {
	if t.checkVersion() {
		ts.Update <- t
	}
}

func (ts *Templates) hasTemplate(view string) bool {
	_, ok := defaultTemplates.Cache[view]
	return ok
}

func (ts *Templates) addTemplate(t *Template) {
	defaultTemplates.Lock.Lock()
	defaultTemplates.Cache[t.Source.Identity] = t
	defaultTemplates.Lock.Unlock()
}

func (ts *Templates) getTemplate(view string) *Template {
	defaultTemplates.Lock.RLock()
	defer defaultTemplates.Lock.RUnlock()

	if t, ok := ts.Cache[view]; ok {
		return t
	}
	return nil
}

func EmptyTemplate() *Template {
	return &Template{
		Lock:          &sync.Mutex{},
		LastParseTime: time.Now(),
	}
}

type Template struct {
	Tr            *Tree
	Lock          *sync.Mutex
	LastParseTime time.Time
	Type          int
	Source        *Source
}

func (t *Template) parse(s *Source) (err error) {
	t.Source = s
	var stream *TokenStream
	stream, err = NewLexer().Tokenize(t.Source)
	if err == nil {
		filter := &TokenFilter{Tr: &Tree{}}
		t.Tr = filter.Filter(stream)
	}
	return errors.WithStack(err)
}

func (t *Template) ParseFile(path string) error {
	t.Type = TPL_TYPE_FILE
	return t.parse(NewSourceFile(path))
}

func (t *Template) ParseString(tpl string) error {
	t.Type = TPL_TYPE_STRING
	return t.parse(NewSource(tpl))
}

func (t *Template) Execute(data ...any) []byte {
	ds, _ := json.MarshalIndent(t.Tr, "tpl", "····")
	return ds
}

func (t *Template) update() error {
	if t.Type == TPL_TYPE_FILE {
		t.Lock.Lock()
		defer t.Lock.Unlock()
		return t.parse(NewSourceFile(t.Source.Identity))
	}
	return nil
}

func (t *Template) checkVersion() bool {
	if t.Type == TPL_TYPE_FILE {
		if info, err := os.Stat(t.Source.Identity); err == nil {
			if info.ModTime().After(t.LastParseTime) {
				return true
			}
		}
	}
	return false
}

func RenderString(w io.Writer, view string, data ...any) error {
	identity := abstract([]byte(view))

	if !defaultTemplates.hasTemplate(identity) {
		t := EmptyTemplate()
		if err := t.ParseString(view); err != nil {
			return errors.WithStack(err)
		}
		defaultTemplates.addTemplate(t)
	}

	if t := defaultTemplates.getTemplate(identity); t != nil {
		if _, err := w.Write(t.Execute(data)); err != nil {
			return errors.WithStack(err)
		}
		return nil
	}

	return errors.New(fmt.Sprintf("Template\n %s \nparsed error", view))
}

func Render(w io.Writer, viewPath string, data ...any) error {

	if !defaultTemplates.hasTemplate(viewPath) {
		t := EmptyTemplate()
		if err := t.ParseFile(viewPath); err != nil {
			return errors.WithStack(err)
		}
		defaultTemplates.addTemplate(t)
	}

	if t := defaultTemplates.getTemplate(viewPath); t != nil {
		if _, err := w.Write(t.Execute(data)); err != nil {
			return errors.WithStack(err)
		}
		return nil
	}

	return errors.New(fmt.Sprintf("Template %s parsed error", viewPath))
}
