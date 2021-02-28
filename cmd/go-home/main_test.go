package main

import (
	"os"
	"testing"
)

func Test_main(t *testing.T) {
	os.Setenv("URL", "0.0.0.0")
	os.Setenv("PORT", "5575")
	main()
}
