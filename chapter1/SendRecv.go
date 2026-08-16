package pi

func Send(ch Name, payload Name) {
	ch <- payload
}

func Recv(in Name, out Name) {
	z := <-in
	Send(out, z)
}
