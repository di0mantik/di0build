package model

type Status int

const (
	None Status = iota
	Done
	Failed
)

type Phase int

const (
	Packages Phase = iota
	Symlinks
)
