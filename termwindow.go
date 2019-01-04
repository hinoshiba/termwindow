package termwindow

import (
	"github.com/nsf/termbox-go"
	"fmt"
)

func Init() error {
    if err := termbox.Init(); err != nil {
    	return err
    }
    EvKey = make(chan EVKEY)
    Error = make(chan error)
    return nil
}

func Close() error {
    termbox.Close()
	return nil
}

func Start() error {
    termbox.Flush()
    return nil
}

const coldef = termbox.ColorDefault

func drawLine(x, y int, str string) {
	runes := []rune(str)
	for i := 0; i < len(runes); i++ {
		termbox.SetCell(x+i, y, runes[i], termbox.ColorDefault, termbox.ColorDefault)
	}
}

type EVKEY struct {
	Key	termbox.Key
	Ch	rune
}

func Input() error {
	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventError:
			Error<-ev.Err
		case termbox.EventResize:
			termbox.Flush()
		case termbox.EventKey:
			switch ev.Key {
			case termbox.KeyCtrlC:
				panic("ctrl\n")
			case termbox.KeyEsc:
				panic("esc\n")
			default:
				EvKey<-EVKEY{Key:ev.Key,Ch:ev.Ch}
			}
		}
	}
}

var EvKey chan EVKEY
func Echo() {
	for {
		select {
		case ev := <-EvKey:
    			termbox.Clear(coldef, coldef)
    			drawLine(1, 0, string(ev.Ch) + "")
    			termbox.Flush()
		default:
			continue
		}
	}
}

var Error chan error
func Err() {
	for {
		select {
		case err := <-Error:
    			termbox.Clear(coldef, coldef)
    			drawLine(1, 0, fmt.Sprintf("%s",err))
    			termbox.Flush()
		default:
			continue
		}
	}
}
