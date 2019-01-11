package termwindow

import (
	"github.com/nsf/termbox-go"
	"goctx"
	"errors"
	"fmt"
)

type Wsegm struct {
	Title   chan []byte
	body    chan []byte
	zbody   chan []byte
	Msg     chan []byte
	Flush   chan struct{}
	Err     chan error
}

func newWsegm() Wsegm {
	var wsegm Wsegm
	wsegm.Title = make(chan []byte)
	wsegm.body  = make(chan []byte)
	wsegm.zbody = make(chan []byte)
	wsegm.Msg   = make(chan []byte)
	wsegm.Flush = make(chan struct{})
	wsegm.Err   = make(chan error)
	return wsegm
}

func Init() (Wsegm, error) {
	if err := termbox.Init(); err != nil {
		return Wsegm{}, err
	}
	EvKey = make(chan EVKEY)
	return newWsegm(), nil
}

func Close() error {
	termbox.Close()
	return nil
}

func refresh() (int, int) {
    	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	termbox.Flush()
	return termbox.Size()
}

var EvKey chan EVKEY

func Start(wk goctx.Worker, wsegm Wsegm) {
	w, h := refresh()

	var title []byte
//	var body  []byte
	var msg   []byte
	var err   error
	for {
		select {
		case <-wk.RecvCancel():
			wk.Done()
			return
		case <-wsegm.Flush:
			w, h = refresh()
			setTitle(w, title)
			setMsg(w, h, msg)
		case title = <-wsegm.Title:
			setTitle(w, title)
		//case body  = <-wsegm.body:
	//		setBody(w, h, title)
		case msg   = <-wsegm.Msg:
			setMsg(w, h, msg)
		case err   = <-wsegm.Err:
			setError(w, h, err)
		case ev := <-EvKey:
			if ev.Ch == 'a' {
				setError(w, h, errors.New("test"))
			}
    			drawLine(1, 0, string(ev.Ch) + "", termbox.ColorDefault, termbox.ColorDefault)
		}
		termbox.Flush()
	}
	return
}

func setTitle(w int, title []byte) {
	drawLine(0, w, string(title), termbox.ColorDefault, termbox.ColorCyan)
}

func setMsg(w int, h int, msg []byte) {
	if h < 1 {
		return
	}
	drawLine(h-1, w, string(msg), termbox.ColorDefault, termbox.ColorCyan)
}

func setError(w int, h int, err error) {
	if h < 1 {
		return
	}
	errstr := fmt.Sprintf("Error : %s", err)
	drawLine(h-1, w, errstr, termbox.ColorRed, termbox.ColorCyan)
}

func drawLine(y int, w int,  str string, fg, bg termbox.Attribute) {
	runes := []rune(str)
	for i := 0; i < len(runes); i++ {
		termbox.SetCell(i, y, runes[i], fg, bg)
	}
	if len(runes) >= w {
		return
	}
	var space rune
	for i := w - len(runes); i >= 0; i-- {
		termbox.SetCell(i + len(runes), y, space, fg, bg)
	}
}

type EVKEY struct {
	Key	termbox.Key
	Ch	rune
}

func Input(wk goctx.Worker, flash chan struct{}) {
	for {
		select {
		case <-wk.RecvCancel():
			wk.Done()
			return
		default:
		}
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventError:
			Error<-ev.Err
		case termbox.EventResize:
			var ret struct{}
			flash<-ret
		case termbox.EventKey:
			switch ev.Key {
			case termbox.KeyCtrlC:
				wk.Cancel()
				wk.Done()
				return
			case termbox.KeyEsc:
				wk.Cancel()
				wk.Done()
				return
			default:
				EvKey<-EVKEY{Key:ev.Key,Ch:ev.Ch}
			}
		}
	}
}



var Error chan error
func Err() {
	for {
		select {
		case <-Error:
    			termbox.Flush()
		default:
			continue
		}
	}
}
