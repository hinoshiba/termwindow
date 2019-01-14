goctx
===

* Make it possible to cancel from goroutine.

## owner
```
func main() {
	own := goctx.NewOwner()
	go gorun(own.NewWorker())
	go gorun(own.NewWorker())
	go gorun(own.NewWorker())
}
```

## routine
```
func gorun(wk goctx.Worker) {
	for {
		select {
		case <-wk.RecvCancel():
			return
		default:
		}
		//yourroutine
		wk.Cancel() //call the withcancel at owner.
	}
}
```
