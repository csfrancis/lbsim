package main

type Algorithm interface {
	Select(peers []*Unicorn) *Unicorn
	ProcessResponse(response *Response)
}
