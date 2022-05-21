package main

type BoxfileDescriber struct {
	File Boxfile
}

func (e *BoxfileDescriber) Describe() {
	boxfile.Load("./Boxfile")	
}