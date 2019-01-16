package termwindow

type Lines struct {
	Active	int
	Data	WinData
}

type WinData struct {
	Title		[]byte
	Body		[][]byte
	Ids		[]int
}

func (self *Lines) Append(id int, data []byte) {
	self.Data.Ids = append(self.Data.Ids, id)
	self.Data.Body = append(self.Data.Body, data)
}

func (self *Lines) GetData(act int) (int, []byte){
	if len (self.Data.Ids) - 1 <  act || len (self.Data.Body) - 1 < act {
		return 0, []byte{}
	}
	return self.Data.Ids[act], self.Data.Body[act]
}

func (self *Lines) MvInc() int {
	if len(self.Data.Body) - 1 > self.Active {
		self.Active++
	}
	return self.Active
}

func (self *Lines) MvDec() int {
	if 0 < self.Active {
		self.Active--
	}
	return self.Active
}

func (self *Lines) MvTop() int {
	self.Active = 0
	return self.Active
}

func (self *Lines) MvBottom() int {
	self.Active = len(self.Data.Body) - 1
	return self.Active
}
