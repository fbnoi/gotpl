package template

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sync"
	"time"

	"github.com/pkg/errors"
)

var (
	defaultTemplates = &Templates{
		Cache:  make(map[string]*Template),
		Lock:   &sync.RWMutex{},
		Update: make(chan *Template),
	}
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
			for _, Template := range ts.Cache {
				go ts.checkVersion(Template)
			}
		case t := <-ts.Update:
			t.Update()
		}
	}
}

func (ts *Templates) checkVersion(t *Template) {
	if info, err := os.Stat(t.Name); err == nil {
		if info.ModTime().After(t.LastParseTime) {
			ts.Update <- t
		}
	}
}

func (ts *Templates) hasTemplate(view string) bool {
	_, ok := defaultTemplates.Cache[view]
	return ok
}

func (ts *Templates) addTemplate(view string) error {

	t := &Template{Lock: &sync.Mutex{}}
	if err := t.Parse(view); err != nil {
		return err
	}
	defaultTemplates.Lock.Lock()
	defer defaultTemplates.Lock.Unlock()

	defaultTemplates.Cache[view] = t

	return nil
}

func (ts *Templates) getTemplate(view string) *Template {
	defaultTemplates.Lock.RLock()
	defer defaultTemplates.Lock.RUnlock()

	if t, ok := ts.Cache[view]; ok {
		return t
	}
	return nil
}

type Template struct {
	Name          string
	Tr            *Tree
	Lock          *sync.Mutex
	LastParseTime time.Time
}

func (t *Template) readFile(path string) (string, error) {
	fs, err := ioutil.ReadFile(path)
	if err != nil {
		return "", errors.WithStack(err)
	}
	return string(fs), nil
}

func (t *Template) Parse(tpl string) (err error) {
	var stream *TokenStream
	stream, err = NewLexer().Tokenize(tpl)
	if err == nil {
		filter := &TokenFilter{Tr: &Tree{}}
		t.Tr = filter.Filter(stream)
	}

	return errors.WithStack(err)
}

func (t *Template) Update() error {
	t.Lock.Lock()
	defer t.Lock.Unlock()
	return t.Parse(t.Name)
}

func (t *Template) Execute(data ...any) []byte {
	return nil
}

func Render(w io.Writer, view string, data ...any) error {

	if !defaultTemplates.hasTemplate(view) {
		if err := defaultTemplates.addTemplate(view); err != nil {
			return errors.WithStack(err)
		}
	}

	if t := defaultTemplates.getTemplate(view); t != nil {
		if _, err := w.Write(t.Execute(data)); err != nil {
			return errors.WithStack(err)
		}
		return nil
	}

	return errors.New(fmt.Sprintf("Template %s parsed error", view))
}
