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
		cache: make(map[string]*template),
		lock:  &sync.RWMutex{},
		up:    make(chan *template),
	}
)

type Config struct {
}

type Templates struct {
	cache map[string]*template
	lock  *sync.RWMutex

	up chan *template
}

func (ts *Templates) Watch() {
	timer := time.NewTicker(time.Millisecond * 1000)
	for {
		select {
		case <-timer.C:
			for _, template := range ts.cache {
				go ts.checkVersion(template)
			}
		case t := <-ts.up:
			t.update()
		}
	}
}

func (ts *Templates) checkVersion(t *template) {
	if info, err := os.Stat(t.name); err == nil {
		if info.ModTime().After(t.lastParseTime) {
			ts.up <- t
		}
	}
}

func (ts *Templates) hasTemplate(view string) bool {
	_, ok := defaultTemplates.cache[view]
	return ok
}

func (ts *Templates) addTemplate(view string) error {

	t := &template{lock: &sync.Mutex{}}
	if err := t.parse(view); err != nil {
		return err
	}
	defaultTemplates.lock.Lock()
	defer defaultTemplates.lock.Unlock()

	defaultTemplates.cache[view] = t

	return nil
}

func (ts *Templates) getTemplate(view string) *template {
	defaultTemplates.lock.RLock()
	defer defaultTemplates.lock.RUnlock()

	if t, ok := ts.cache[view]; ok {
		return t
	}
	return nil
}

type template struct {
	name          string
	tr            *Tree
	lock          *sync.Mutex
	lastParseTime time.Time
}

func (t *template) readFile(path string) (string, error) {
	fs, err := ioutil.ReadFile(path)
	if err != nil {
		return "", errors.WithStack(err)
	}
	return string(fs), nil
}

func (t *template) parse(tpl string) (err error) {
	var stream *TokenStream
	stream, err = Lexer().Tokenize(tpl)
	if err == nil {
		filter := &TokenFilter{Tr: &Tree{}}
		t.tr = filter.Filter(stream)
	}

	return errors.WithStack(err)
}

func (t *template) update() error {
	t.lock.Lock()
	defer t.lock.Unlock()
	return t.parse(t.name)
}

func (t *template) execute(data ...any) []byte {
	return nil
}

func Render(w io.Writer, view string, data ...any) error {

	if !defaultTemplates.hasTemplate(view) {
		if err := defaultTemplates.addTemplate(view); err != nil {
			return errors.WithStack(err)
		}
	}

	if t := defaultTemplates.getTemplate(view); t != nil {
		if _, err := w.Write(t.execute(data)); err != nil {
			return errors.WithStack(err)
		}
		return nil
	}

	return errors.New(fmt.Sprintf("template %s parsed error", view))
}
