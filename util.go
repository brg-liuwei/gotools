package gotools

func assert(expression bool) {
	if !expression {
		panic("assert failed")
	}
}
