package debug

func Assert(cond bool) {
	if !cond {
		panic("condition not met")
	}
}

func Check(err error) {
	if err != nil {
		panic(err)
	}
}
