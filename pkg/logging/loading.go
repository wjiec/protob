package logging

import (
	"os"
	"time"

	"github.com/briandowns/spinner"
)

const (
	loadingCharacterIndex    = 14
	loadingAnimationDuration = 80 * time.Millisecond
)

var (
	bar           = spinner.New(spinner.CharSets[loadingCharacterIndex], loadingAnimationDuration)
	loadingLogger = NewLogger(os.Stdout)
)

func init() {
	bar.HideCursor = true
	bar.Prefix = " "
}

type Bar struct{}

func (b *Bar) Text(text string) {
	bar.Suffix = " " + text
}

func (b *Bar) Success(format string, args ...interface{}) {
	bar.Stop()
	loadingLogger.Success(format, args...)
}

func (b *Bar) Fatal(format string, args ...interface{}) {
	bar.Stop()
	loadingLogger.Fatal(format, args...)
}

func (b *Bar) Error(err error) {
	if err != nil {
		b.Fatal(err.Error())
	}
}

func Loading(text string, action func(*Bar)) {
	bar.Suffix = " " + text
	bar.Start()
	action(&Bar{})
	bar.Stop()
}
