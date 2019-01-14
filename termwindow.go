package termwindow

import (
	"github.com/nsf/termbox-go"
	"github.com/hinoshiba/goctx"
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

type windows struct {
	menu	window
	body	window
}

type window struct {
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
	var wins	windows
	var err		error

	for {
		select {
		case <-wk.RecvCancel():
			wk.Done()
			return
		case <-Flush:
			w, h = refresh()
			drawTitle(w, &title)
			drawMsg(w, h, &msg)
			drawWindows(w, h, &wins)
		case title = <-Title:
			drawTitle(w, &title)
		case wins.menu.Data = <-Menu:
			drawWindows(w, h, &wins)
		case aline := <-ActiveLine:
			attach := getAttached(&wins)
			attach.Active = aline
			drawWindows(w, h, &wins)
		case wins.body.Data = <-Body:
			wins.body.Active = 0
			wins.body.Head = 0
			drawWindows(w, h, &wins)
		case msg  = <-Msg:
			drawMsg(w, h, &msg)
		case err  = <-Err:
			drawError(w, h, &err)
		}
		termbox.Flush()
	}
	return
}

func getAttached(wins *windows) *window {
	if wins.body.Data != nil {
		return &wins.body
	}
	return &wins.menu
}

func drawTitle(w int, title *[]byte) {
	drawLine(WINDOW_TOP, w, string(*title), CL_TITLE_FG, CL_TITLE_BG)
}

func drawMsg(w int, h int, msg *[]byte) {
	if h < TITLE_HEIGHT {
		return
	}
	drawLine(h - MSG_HEIGHT, w, string(*msg), CL_MSG_FG, CL_MSG_BG)
}

func drawError(w int, h int, err *error) {
	if h < TITLE_HEIGHT {
		return
	}
	if *err == nil {
		return
	}
	errstr := fmt.Sprintf("Error : %s", *err)
	drawLine(h - MSG_HEIGHT, w, errstr, CL_ERR_FG, CL_ERR_BG)
}

func drawWindows(w int, h int, wins *windows) {
	if wins.body.Data != nil {
		hbs := (h - 1) / 2
		hbe := h - hbs - MSG_HEIGHT - 1
		drawWindow(w, 0, hbs - 1, &wins.menu)
		drawWindow(w, hbs, hbe, &wins.body)
		return
	}
	drawWindow(w, 0, h -  MSG_HEIGHT - TITLE_HEIGHT , &wins.menu)
}

func drawWindow(w int, hs int, max_line int, win *window) {
	if win.Active < 0 {
		errp("A actice value outside the range was specified at the window.")
		return
	}
	if len(win.Data) - 1 < win.Active {
		errp("A actice value outside the range was specified at the window.")
		return
	}
	if len(win.Data) - 1 < win.Head {
		errp("A head value outside the range was specified at the window.")
		return
	}

	var cnt int = hs

	if max_line <= win.Active - win.Head {
		win.Head = win.Active - max_line + 1
	}
	if win.Head >= win.Active {
		win.Head = win.Active
	}
	if win.Head < 0 {
		return
	}

	act := win.Active - win.Head
	limit := max_line + hs
	//msgp(fmt.Sprintf("head :%v, Active:%v, MaxLine:%v, ColorBG:%v, ColorFG:%v",win.Head, win.Active,max_line, CL_MENUACT_BG, CL_MENUACT_FG))
	for i, d := range win.Data[win.Head:] {
		if ! (cnt < limit) {
			return
		}
		if i == act {
			drawLine(cnt + TITLE_HEIGHT, w, string(d),
						CL_MENUACT_FG, CL_MENUACT_BG)
		} else {
			drawLine(cnt + TITLE_HEIGHT, w, string(d),
						CL_MENU_FG, CL_MENU_BG)
		}
		cnt++
	}

	for ; cnt < limit; cnt++ {
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
				errp("canceld")
				wk.Cancel()
				wk.Done()
				return
			case termbox.KeyEsc:
				msgp("escaped")
				wk.Cancel()
				wk.Done()
				return
			default:
				Key<-EVKEY{Key:ev.Key,Ch:ev.Ch}
			}
		}
	}
}

func SetTitle(str string) {
	go func(){
		Title<-[]byte(str)
	}()
}

func SetMsg(str string) {
	go func(){
		Msg<-[]byte(str)
	}()
}

func SetErrStr(str string) {
	go func(){
		err := errors.New(str)
		Err<-err
	}()
}

func SetErr(err error) {
	go func(){
		Err<-err
	}()
}

func SetActiveLine(i int) {
	go func(){
		ActiveLine<-i
	}()
}

func SetMenu(b [][]byte) {
	go func(){
		Menu<-b
	}()
}

func SetBody(b [][]byte) {
	go func(){
		Body<-b
	}()
}

func UnsetBody() {
	go func(){
		Body<-nil
	}()
}


func msgp(msg string) {
	go func() {
		Msg <- []byte(msg)
	}()
}

func errp(msg string) {
	err := errors.New(msg)
	go func() {
		Err <- err
	}()
}
