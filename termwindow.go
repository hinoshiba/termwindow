package termwindow

import (
	"github.com/nsf/termbox-go"
	"goctx"
	"errors"
	"fmt"
)

const (
	WINDOW_TOP int = 0
	TITLE_HEIGHT int = 1
	MSG_HEIGHT int = 1

	//TITLE_BG_COLOR = termbox.Attribute(uint16(
	//TITLE_FG_COLOR = termbox.Attribute(uint16(
	//MSG_BG_COLOR = termbox.Attribute(uinti16(
	//MSG_FG_COLOR = termbox.Attribute(uinti16(
	//ERR_BG_COLOR = termbox.Attribute(uinti16(
	//ERR_FG_COLOR = termbox.Attribute(uinti16(
)

type EVKEY struct {
	Key	termbox.Key
	Ch	rune
}

type MenuPropary struct {
	Active	int//tba uint32
	Head	int//tba uint32
	Data	[][]byte
}

var (
	Key		chan EVKEY
	Title		chan []byte
	Menu		chan [][]byte
	ActiveMenu	chan int
	Body		chan [][]byte
	Msg		chan []byte
	Flush		chan struct{}
	Err		chan error
)

func Init() error {
	if err := termbox.Init(); err != nil {
		return err
	}
	Key = make(chan EVKEY)

	Title      = make(chan []byte)
	Menu       = make(chan [][]byte)
	ActiveMenu = make(chan int)
	Body       = make(chan [][]byte)
	Msg        = make(chan []byte)
	Flush      = make(chan struct{})
	Err        = make(chan error)
	return nil
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



func Start(wk goctx.Worker) {
	w, h := refresh()

	var title	[]byte
	var msg		[]byte
	var menu	MenuPropary
	var err		error

	for {
		select {
		case <-wk.RecvCancel():
			wk.Done()
			return
		case <-Flush:
			w, h = refresh()
			setTitle(w, title)
			setMsg(w, h, msg)
			setMenu(w, h, &menu)
		case title = <-Title:
			setTitle(w, title)
		case menu.Data  = <-Menu:
			menu.Active = 0
			menu.Head = 0
			setMenu(w, h, &menu)
		case menu.Active = <-ActiveMenu:
			setMenu(w, h, &menu)
		case msg  = <-Msg:
			setMsg(w, h, msg)
		case err  = <-Err:
			setError(w, h, err)
		case ev := <-Key:
			if ev.Ch == 'a' {
				setError(w, h, errors.New("test"))
			}
			if ev.Ch == 'j' {
				go func() {
					ActiveMenu<-menu.Active + 1
				}()
			}
			if ev.Ch == 'k' {
				go func() {
					ActiveMenu<-menu.Active - 1
				}()
			}
			if ev.Ch == 'G' {
				go func() {
					ActiveMenu<-len(menu.Data) - 1
				}()
			}
			if ev.Ch == 'g' {
				go func() {
					ActiveMenu<-0
				}()
			}
    			drawLine(1, 0, string(ev.Ch) + "", termbox.ColorDefault, termbox.ColorDefault)
		}
		termbox.Flush()
	}
	return
}

func setTitle(w int, title []byte) {
	drawLine(WINDOW_TOP, w, string(title), termbox.ColorDefault, termbox.ColorCyan)
}

func setMsg(w int, h int, msg []byte) {
	if h < TITLE_HEIGHT {
		return
	}
	drawLine(h - MSG_HEIGHT, w, string(msg), termbox.ColorDefault, termbox.ColorCyan)
}

func setMenu(w int, h int, menu *MenuPropary) {
	if menu.Active < 0 {
		Errp("A actice value outside the range was specified at the menu.")
		return
	}
	if len(menu.Data) - 1 < menu.Active {
		Errp("A actice value outside the range was specified at the menu.")
		return
	}
	if len(menu.Data) - 1 < menu.Head {
		Errp("A head value outside the range was specified at the menu.")
		return
	}

	max_line := h - MSG_HEIGHT - TITLE_HEIGHT
	cnt := 0

	if max_line <= menu.Active - menu.Head {
		menu.Head = menu.Active - max_line + 1
	}
	if menu.Head >= menu.Active {
		menu.Head = menu.Active
	}

	if menu.Head < 0 {
		return
	}

	act := menu.Active - menu.Head
	Msgp(fmt.Sprintf("head :%v, Active:%v, MaxLine:%v",menu.Head, menu.Active,max_line))
	for i, d := range menu.Data[menu.Head:] {
		if i == act {
			drawLine(cnt + TITLE_HEIGHT, w, string(d), termbox.ColorDefault, termbox.ColorRed)
		} else {
			drawLine(cnt + TITLE_HEIGHT, w, string(d), termbox.ColorDefault, termbox.ColorDefault)
		}
		cnt++
		if cnt > max_line {
			Errp(fmt.Sprintf("cnt:%v",cnt))
			return
		}
	}
	for ; cnt <= max_line; cnt++ {
		drawLine(cnt + TITLE_HEIGHT, w, "", termbox.ColorDefault, termbox.ColorDefault)
	}

}

func setError(w int, h int, err error) {
	if h < TITLE_HEIGHT {
		return
	}
	if err == nil {
		return
	}
	errstr := fmt.Sprintf("Error : %s", err)
	drawLine(h - MSG_HEIGHT, w, errstr, termbox.ColorRed, termbox.ColorCyan)
}

func Msgp(msg string) {
	go func() {
		Msg <- []byte(msg)
	}()
}

func Errp(msg string) {
	err := errors.New(msg)
	go func() {
		Err <- err
	}()
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

func Input(wk goctx.Worker) {
	for {
		select {
		case <-wk.RecvCancel():
			wk.Done()
			return
		default:
		}
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventError:
			Err <- ev.Err
		case termbox.EventResize:
			var ret struct{}
			Flush <- ret
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
				Key<-EVKEY{Key:ev.Key,Ch:ev.Ch}
			}
		}
	}
}
