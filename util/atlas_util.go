package util

func CheckFatalError(err error) {
	if err != nil {
		panic(err)
	}
}
