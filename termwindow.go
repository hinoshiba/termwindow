package termwindow

import (
	"github.com/nsf/termbox-go"
	"github.com/hinoshiba/goctx"
	"errors"
	"fmt"
	"strings"
//	"strconv"
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

const (
	KeyCtrlTilde      termbox.Key = 0x00
	KeyCtrl2          termbox.Key = 0x00
	KeyCtrlSpace      termbox.Key = 0x00
	KeyCtrlA          termbox.Key = 0x01
	KeyCtrlB          termbox.Key = 0x02
	KeyCtrlC          termbox.Key = 0x03
	KeyCtrlD          termbox.Key = 0x04
	KeyCtrlE          termbox.Key = 0x05
	KeyCtrlF          termbox.Key = 0x06
	KeyCtrlG          termbox.Key = 0x07
	KeyBackspace      termbox.Key = 0x08
	KeyCtrlH          termbox.Key = 0x08
	KeyTab            termbox.Key = 0x09
	KeyCtrlI          termbox.Key = 0x09
	KeyCtrlJ          termbox.Key = 0x0A
	KeyCtrlK          termbox.Key = 0x0B
	KeyCtrlL          termbox.Key = 0x0C
	KeyEnter          termbox.Key = 0x0D
	KeyCtrlM          termbox.Key = 0x0D
	KeyCtrlN          termbox.Key = 0x0E
	KeyCtrlO          termbox.Key = 0x0F
	KeyCtrlP          termbox.Key = 0x10
	KeyCtrlQ          termbox.Key = 0x11
	KeyCtrlR          termbox.Key = 0x12
	KeyCtrlS          termbox.Key = 0x13
	KeyCtrlT          termbox.Key = 0x14
	KeyCtrlU          termbox.Key = 0x15
	KeyCtrlV          termbox.Key = 0x16
	KeyCtrlW          termbox.Key = 0x17
	KeyCtrlX          termbox.Key = 0x18
	KeyCtrlY          termbox.Key = 0x19
	KeyCtrlZ          termbox.Key = 0x1A
	KeyEsc            termbox.Key = 0x1B
	KeyCtrlLsqBracket termbox.Key = 0x1B
	KeyCtrl3          termbox.Key = 0x1B
	KeyCtrl4          termbox.Key = 0x1C
	KeyCtrlBackslash  termbox.Key = 0x1C
	KeyCtrl5          termbox.Key = 0x1D
	KeyCtrlRsqBracket termbox.Key = 0x1D
	KeyCtrl6          termbox.Key = 0x1E
	KeyCtrl7          termbox.Key = 0x1F
	KeyCtrlSlash      termbox.Key = 0x1F
	KeyCtrlUnderscore termbox.Key = 0x1F
	KeySpace          termbox.Key = 0x20
	KeyBackspace2     termbox.Key = 0x7F
	KeyCtrl8          termbox.Key = 0x7F
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
	Data		WinData
}

var (
	Key		chan EVKEY
	Title		chan []byte
	Menu		chan WinData
	ActiveLine	chan int
	Body		chan WinData
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
	Menu       = make(chan WinData)
	ActiveLine = make(chan int)
	Body       = make(chan WinData)
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
	defer wk.Done()
	w, h := refresh()

	var title	[]byte
	var msg		[]byte
	var wins	windows
	var err		error

	for {
		select {
		case <-wk.RecvCancel():
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
	if wins.body.Data.Body != nil {
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
	if wins.body.Data.Body != nil {
		hbs := (h - 1) / 2
		hbe := h - hbs - MSG_HEIGHT - 1
		drawWindow(w, 0, hbs, &wins.menu)
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
	if len(win.Data.Body) - 1 < win.Active {
		errp("A actice value outside the range was specified at the window.")
		return
	}
	if len(win.Data.Body) - 1 < win.Head {
		errp("A head value outside the range was specified at the window.")
		return
	}

	if win.Data.Title == nil {
		if max_line <= win.Active - win.Head {
			win.Head = win.Active - max_line + 1
		}
	}else{
		if max_line - 1 <= win.Active - win.Head {
			win.Head = win.Active - max_line + 2
		}
	}
	if win.Head >= win.Active {
		win.Head = win.Active
	}
	if win.Head < 0 {
		return
	}


	var cnt int = hs

	act := win.Active - win.Head
	limit := max_line + hs
	if ! (cnt < limit) {
		return
	}
	if win.Data.Title != nil {
		drawLine(cnt + TITLE_HEIGHT, w, string(win.Data.Title),
					CL_TITLE_FG, CL_TITLE_BG)
	//	msgp(fmt.Sprintf("cnt:%v, head :%v, Active:%v, hs:%v, act:%v, MaxLine:%v, ColorBG:%v, ColorFG:%v",cnt, win.Head, win.Active, hs, act, max_line, CL_MENUACT_BG, CL_MENUACT_FG))
		cnt++
	}

	for i, d := range win.Data.Body[win.Head:] {
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
		drawLine(cnt + TITLE_HEIGHT, w, " ", CL_MENU_FG, CL_MENU_BG)
	}
}

func drawLine(y int, w int,  str string, fg, bg termbox.Attribute) {
	var wrote int
	var c_flag = false
	var c_txt string
	var c_fg termbox.Attribute
	var c_bg termbox.Attribute

	runes := []rune(str)

	c_fg = fg
	c_bg = bg
	for _, rune := range runes {
		if w < wrote {
			return
		}
		// [38;5;97;48;5;107mtest [0;00m [0m  [38;5;26;48;5;178mtest [0;00m [0m
		if "\x1b" == string(rune) {
			c_flag = true
		}
		if c_flag {
			if "m" == string(rune) {
				c_flag = false
				c_cls := strings.Split(c_txt, ";")
				if len(c_cls) < 6 {
					c_fg = fg
					c_bg = bg
					continue
				} else {
				c_fg = termbox.ColorBlack
				c_bg = termbox.ColorWhite
				}/* else {
					u_bg, _ := strconv.ParseUint(c_cls[2], 16, 16)
					c_bg = termbox.Attribute(u_bg)
					u_fg, _ := strconv.ParseUint(c_cls[5], 16, 16)
					c_fg = termbox.Attribute(u_fg)
				}
				*/
				c_txt = ""
			}
			c_txt += string(rune)
			continue
		}
		termbox.SetCell(wrote, y, rune, c_fg, c_bg)
		if len([]byte(string(rune))) > 1 {
			wrote += 2
			continue
		}
		wrote++
	}

	var space rune
	for ; w >= wrote; wrote++ {
		termbox.SetCell(wrote, y, space, fg, bg)
	}
}


func Input(wk goctx.Worker) {
	defer wk.Done()
	for {
		select {
		case <-wk.RecvCancel():
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
				return
			case termbox.KeyEsc:
				msgp("escaped")
				wk.Cancel()
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

func SetMsg(s string, msg ...interface{}) {
	str := fmt.Sprintf(s , msg...)
	go func(){
		Msg<-[]byte(str)
	}()
}

func SetErrStr(s string, msg ...interface{}) {
	err := errors.New(fmt.Sprintf(s , msg...))
	go func(){
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

func SetMenu(b WinData) {
	go func(){
		Menu<-b
	}()
}

func SetBody(b WinData) {
	go func(){
		Body<-b
	}()
}

func UnsetBody() {
	go func(){
		Body<-WinData{}
	}()
}

func ReFlush() {
	go func(){
		var ret struct{}
		Flush <- ret
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
