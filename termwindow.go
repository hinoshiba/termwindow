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

	CL_TITLE_BG = termbox.Attribute(uint16(0x12))
	CL_TITLE_FG = termbox.Attribute(uint16(0xFD))
	CL_MSG_BG = termbox.Attribute(uint16(0x12))
	CL_MSG_FG = termbox.Attribute(uint16(0xFD))
	CL_ERR_BG = termbox.Attribute(uint16(0x12))
	CL_ERR_FG = termbox.Attribute(uint16(0xC5))
	CL_MENU_BG = termbox.ColorDefault
	CL_MENU_FG = termbox.ColorDefault
	CL_MENUACT_BG = termbox.Attribute(uint16(0x08))
	CL_MENUACT_FG = termbox.Attribute(uint16(0x17))
)

type EVKEY struct {
	Key	termbox.Key
	Ch	rune
}

type Windows struct {
	menu	Window
	body	Window
}
type Window struct {
	Active		int//tba uint32
	Head		int//tba uint32
	Data		[][]byte
}

var (
	Key		chan EVKEY
	Title		chan []byte
	Menu		chan [][]byte
	ActiveLine	chan int
	Body		chan [][]byte
	Msg		chan []byte
	Flush		chan struct{}
	Err		chan error
)

func Init() error {
	if err := termbox.Init(); err != nil {
		return err
	}
	termbox.SetOutputMode(termbox.Output256)
	Key = make(chan EVKEY)

	Title      = make(chan []byte)
	Menu       = make(chan [][]byte)
	ActiveLine = make(chan int)
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
	var windows	Windows
	var err		error

	for {
		select {
		case <-wk.RecvCancel():
			wk.Done()
			return
		case <-Flush:
			w, h = refresh()
			setTitle(w, &title)
			setMsg(w, h, &msg)
			setWindows(w, h, &windows)
		case title = <-Title:
			setTitle(w, &title)
		case windows.menu.Data  = <-Menu:
			windows.menu.Active = 0
			windows.menu.Head = 0
			setWindows(w, h, &windows)
		case aline := <-ActiveLine:
			attach := getAttach(&windows)
			attach.Active = aline
			setWindows(w, h, &windows)
		case windows.body.Data = <-Body:
			windows.body.Active = 0
			windows.body.Head = 0
			setWindows(w, h, &windows)
		case msg  = <-Msg:
			setMsg(w, h, &msg)
		case err  = <-Err:
			setError(w, h, &err)
		case ev := <-Key:
			if ev.Ch == 'j' {
				go func() {
					attach := getAttach(&windows)
					ActiveLine<- attach.Active + 1
				}()
			}
			if ev.Ch == 'k' {
				go func() {
					attach := getAttach(&windows)
					ActiveLine<- attach.Active - 1
				}()
			}
			if ev.Ch == 'G' {
				go func() {
					attach := getAttach(&windows)
					ActiveLine<-len(attach.Data) - 1
				}()
			}
			if ev.Ch == 'g' {
				go func() {
					ActiveLine<-0
				}()
			}
			if ev.Ch == 'o' {
				go func() {
				var testbody [][]byte
				for i:=0; i<=100; i++ {
					testbody = append(testbody, []byte(fmt.Sprintf("body_line:%v", i)))
				}
				Body<-testbody
				}()
			}
			if ev.Ch == 'c' {
				go func() {
				Body<-nil
				}()
			}
		}
		termbox.Flush()
	}
	return
}

func getAttach(windows *Windows) *Window {
	if windows.body.Data != nil {
		return &windows.body
	}
	return &windows.menu
}

func setTitle(w int, title *[]byte) {
	drawLine(WINDOW_TOP, w, string(*title), CL_TITLE_FG, CL_TITLE_BG)
}

func setMsg(w int, h int, msg *[]byte) {
	if h < TITLE_HEIGHT {
		return
	}
	drawLine(h - MSG_HEIGHT, w, string(*msg), CL_MSG_FG, CL_MSG_BG)
}

func setError(w int, h int, err *error) {
	if h < TITLE_HEIGHT {
		return
	}
	if *err == nil {
		return
	}
	errstr := fmt.Sprintf("Error : %s", *err)
	drawLine(h - MSG_HEIGHT, w, errstr, CL_ERR_FG, CL_ERR_BG)
}

func setWindows(w int, h int, windows *Windows) {
	if windows.body.Data != nil {
		hbs := h / 2
		hbe := h - hbs
		setWindow(w, 0, hbs, &windows.menu)
		setWindow(w, hbs, hbe, &windows.body)
		return
	}
	setWindow(w, 0, h, &windows.menu)
}

func setWindow(w int, hs int, he int, window *Window) {
	if window.Active < 0 {
		window.Active = 0
		Errp("A actice value outside the range was specified at the window.")
		return
	}
	if len(window.Data) - 1 < window.Active {
		window.Active = len(window.Data) - 1
		Errp("A actice value outside the range was specified at the window.")
		return
	}
	if len(window.Data) - 1 < window.Head {
		window.Head = len(window.Data)
		Errp("A head value outside the range was specified at the window.")
		return
	}

	max_line := he - MSG_HEIGHT - TITLE_HEIGHT
	var cnt int = hs

	if max_line <= window.Active - window.Head {
		window.Head = window.Active - max_line + 1
	}
	if window.Head >= window.Active {
		window.Head = window.Active
	}
	if window.Head < 0 {
		return
	}

	act := window.Active - window.Head
	Msgp(fmt.Sprintf("head :%v, Active:%v, MaxLine:%v, ColorBG:%v, ColorFG:%v",window.Head, window.Active,max_line, CL_MENUACT_BG, CL_MENUACT_FG))
	for i, d := range window.Data[window.Head:] {
		if i == act {
			drawLine(cnt + TITLE_HEIGHT, w, string(d),
						CL_MENUACT_FG, CL_MENUACT_BG)
		} else {
			drawLine(cnt + TITLE_HEIGHT, w, string(d),
						CL_MENU_FG, CL_MENU_BG)
		}
		cnt++
		if cnt > max_line + hs {
			return
		}
	}

	max_emp := max_line + hs + 1
	for ; cnt <= max_emp; cnt++ {
		drawLine(cnt + TITLE_HEIGHT, w, "", CL_MENU_FG, CL_MENU_BG)
	}
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
