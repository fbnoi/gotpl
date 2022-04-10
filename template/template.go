package template

import (
	"io"
	"os"
	"sync"
	"time"
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

type template struct {
	name          string
	tr            *Tree
	lastParseTime time.Time
}

func (t *template) Parse(view string) {

}

func (t *template) update() {

}

func (t *template) Execute(data ...any) []byte {
	return nil
}

func Render(w io.Writer, view string, data ...any) {
}
