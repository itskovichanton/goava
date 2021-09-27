package async

func Execute(f func(), async bool) {
	if async {
		go f()
	} else {
		f()
	}
}
