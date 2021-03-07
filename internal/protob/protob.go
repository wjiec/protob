package protob

import (
	"protob/pkg/os/fs"
	"protob/pkg/protobuf"

	"github.com/mitchellh/go-homedir"
)

// Home returns path of the protob home
func Home() string {
	home, _ := homedir.Dir()

	return fs.Join(home, ".protob")
}

// Dependency returns path of the dependencies
func Dependency() string {
	return fs.Join(Home(), "include")
}

// Compiler returns path of the embedded compiler
func Compiler() string {
	return fs.Join(Home(), protobuf.CompilerExecutable)
}

// Temporary returns path of the temporary directory
func Temporary() string {
	return fs.Join(Home(), ".temp")
}
