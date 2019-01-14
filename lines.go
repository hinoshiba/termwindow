package termwindow

type Lines struct {
	Active	int
	Ids	[]int
	Data	[][]byte
}

func (self *Lines) Append(id int, data []byte) {
	self.Ids = append(self.Ids, id)
	self.Data = append(self.Data, data)
}

func (self *Lines) GetData(cnt int) (int, []byte){
	if len (self.Ids) - 1 < cnt || len (self.Data) - 1 < cnt {
		return 0, []byte{}
	}
	return self.Ids[cnt], self.Data[cnt]
}

func (self *Lines) MvInc() int {
	if len(self.Data) > self.Active + 1 {
		self.Active++
	}
	return self.Active
}

func (self *Lines) MvDec() int {
	if 0 <= self.Active - 1 {
		self.Active--
	}
	return self.Active
}

func (self *Lines) MvTop() int {
	self.Active = 0
	return self.Active
}

func (self *Lines) MvBottom() int {
	self.Active = len(self.Data) - 1
	return self.Active
}
