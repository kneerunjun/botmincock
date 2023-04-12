package main

import "reflect"

type ReplyConfig struct {
	Text string
}
type IReply interface {
	Make(*ReplyConfig) IReply
	Text() string
}

type TextReply struct {
	msg string
}

func (tr *TextReply) Make(cfg *ReplyConfig) IReply {
	tr.msg = cfg.Text
	return tr
}
func (tr *TextReply) Text() string {
	return tr.msg
}

type PollReply struct {
	question string
	options  []string
}

func NewReply(ty reflect.Type, attr *ReplyConfig) IReply {
	itf := reflect.New(ty.Elem()).Interface()
	repl, ok := itf.(IReply)
	if !ok {
		return nil
	}
	return repl.Make(attr)
}
