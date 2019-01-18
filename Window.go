package termwindow

import "fmt"

type Window struct {
	Active	int
	Data	WinData
}

type WinData struct {
	Title		[]byte
	Body		[][]byte
	Ids		[]string
}

func (self *Window) SetTitle(s string, msg ...interface{}) {
	self.Data.Title = []byte(fmt.Sprintf(s , msg...))
}

func (self *Window) Append(id string, data []byte) {
	self.Data.Ids = append(self.Data.Ids, id)
	self.Data.Body = append(self.Data.Body, data)
}

func (self *Window) GetData(act int) (string, []byte){
	if len (self.Data.Ids) - 1 <  act || len (self.Data.Body) - 1 < act {
		return "", []byte{}
	}
	return self.Data.Ids[act], self.Data.Body[act]
}

func (self *Window) MvInc() int {
	if self.Data.Body == nil {
		return 0
	}
	if len(self.Data.Body) - 1 > self.Active {
		self.Active++
	}
	return self.Active
}

func (self *Window) MvDec() int {
	if self.Data.Body == nil {
		return 0
	}
	if 0 < self.Active {
		self.Active--
	}
	return self.Active
}

func (self *Window) MvTop() int {
	if self.Data.Body == nil {
		return 0
	}
	self.Active = 0
	return self.Active
}

func (self *Window) MvBottom() int {
	if self.Data.Body == nil {
		return 0
	}
	self.Active = len(self.Data.Body) - 1
	return self.Active
}
