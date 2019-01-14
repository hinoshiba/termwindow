package main

import (
	"termwindow"
	"fmt"
	"github.com/hinoshiba/goctx"
)

func main() {
	own := goctx.NewOwner()

	if err := termwindow.Init(); err != nil {
		return
	}

	go termwindow.Input(own.NewWorker())
	go termwindow.Start(own.NewWorker())
	defer termwindow.Close()

	termwindow.SetTitle("test titile\nasdfadfs")
	termwindow.SetMsg("test Msg")

	var testmenu termwindow.Lines
	for i:=0; i<=100; i++ {
		testmenu.Append(i, []byte(fmt.Sprintf("line:%v", i)))
	}
	termwindow.SetMenu(testmenu.Data)

	wk := own.NewWorker()
	go func(){
		for {
			select {
			case <-wk.RecvCancel():
				wk.Done()
				return
			case ev := <-termwindow.Key:
				if ev.Ch == 'j' {
					termwindow.SetActiveLine(testmenu.MvInc())
				}
				if ev.Ch == 'k' {
					termwindow.SetActiveLine(testmenu.MvDec())
				}
				if ev.Ch == 'G' {
					termwindow.SetActiveLine(testmenu.MvBottom())
				}
				if ev.Ch == 'g' {
					termwindow.SetActiveLine(testmenu.MvTop())
				}
				if ev.Ch == 'o' {
					go func() {
					var testbody [][]byte
					for i:=0; i<=100; i++ {
						testbody = append(testbody, []byte(fmt.Sprintf("body_line:%v", i)))
					}
					termwindow.SetBody(testbody)
					}()
				}
				if ev.Ch == 'c' {
					termwindow.UnsetBody()
				}
			default:
			}
		}
	}()
	own.Wait()
}
