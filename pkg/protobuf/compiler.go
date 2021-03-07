package protobuf

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"protob/pkg/os/fs"
	"strings"
)

var (
	// ErrCompilerNotFound represents unable to found protobuf compiler on system path
	ErrCompilerNotFound = errors.New("protoc: not found")
	// ErrCompilerInvalid represents protobuf compiler not a executable or something error
	ErrCompilerInvalid = errors.New("protoc: invalid executable")
)

// Compiler represents a protobuf compiler
type Compiler struct {
	// Compiler version number
	Version string

	// path of the compiler
	path string
}

// Compile compile protobuf into go file
func (c *Compiler) Compile(target string, runtime *CompilerRuntime) error {
	if out, err := exec.Command(c.path, runtime.Build(target)...).CombinedOutput(); err != nil {
		return errors.New(fmt.Sprintf("%s", out))
	}
	return nil
}

// NewCompiler create a compiler from path
func NewCompiler(path string) (*Compiler, error) {
	if path != "" {
		if ok, _ := fs.IsFile(path); ok {
			if output, err := exec.Command(path, "--version").Output(); err == nil {
				return &Compiler{path: path, Version: strings.TrimSpace(string(output))}, nil
			}
		}
	}

	return nil, ErrCompilerInvalid
}

// NewSystemCompiler create a compiler lookup form system path
func NewSystemCompiler() (*Compiler, error) {
	if path, err := exec.LookPath(CompilerExecutable); err != nil {
		return nil, ErrCompilerNotFound
	} else {
		return NewCompiler(path)
	}
}

const (
	extFast uint8 = 1 << iota
	extFaster
	extSlick
)

// CompilerRuntime represents an runtime for one compile
type CompilerRuntime struct {
	// protobuf dependencies
	dependencies []string

	// whether compile with grpc
	grpc bool

	// any extension to enabled
	extension uint8

	// external argument for protobuf compiler
	arguments []string

	// compiler options: source_relative
	sourceRelative bool

	// output directory
	output string
}

// Build build compile command arguments
func (runtime *CompilerRuntime) Build(target string) []string {
	var args []string
	for _, dependency := range runtime.dependencies {
		args = append(args, "-I", dependency)
	}
	args = append(args, "-I", fs.NormalizePath(filepath.Dir(target)))

	var output string
	switch runtime.extension {
	case extFast:
		output += "gogofast_out="
	case extFaster:
		output += "gogofaster_out="
	case extSlick:
		output += "gogoslick_out="
	}

	if runtime.grpc {
		output += "plugins=grpc,"
	}

	if runtime.sourceRelative {
		output += "paths=source_relative,"
	}

	outputDir := runtime.output
	if outputDir == "" {
		outputDir = fs.NormalizePath(filepath.Dir(target))
	}

	args = append(args, fmt.Sprintf("--%s:%s", strings.TrimRight(output, ","), outputDir))
	args = append(args, target)

	return args
}

// NewCompileRuntime create an runtime for compile by options
func NewCompileRuntime(options ...CompileOption) *CompilerRuntime {
	runtime := &CompilerRuntime{dependencies: []string{}, extension: extSlick, arguments: []string{}}
	for _, option := range options {
		option(runtime)
	}

	if path := os.Getenv("GOPATH"); path != "" {
		runtime.dependencies = append(runtime.dependencies, fs.Join(path, "src"))
	}

	return runtime
}

// CompileOption represents an option to control compile
type CompileOption func(*CompilerRuntime)

// WithGrpc enable grpc extension or not
func WithGrpc(grpc bool) CompileOption {
	return func(runtime *CompilerRuntime) {
		runtime.grpc = grpc
	}
}

// WithExtFaster sets compile extension using gogofast
func WithExtFast(fast bool) CompileOption {
	return func(runtime *CompilerRuntime) {
		if fast {
			runtime.extension = extFast
		}
	}
}

// WithExtFaster sets compile extension using gogofaster
func WithExtFaster(faster bool) CompileOption {
	return func(runtime *CompilerRuntime) {
		if faster {
			runtime.extension = extFaster
		}
	}
}

// WithExtFaster sets compile extension using gogoslick
func WithExtSlick(slick bool) CompileOption {
	return func(runtime *CompilerRuntime) {
		if slick {
			runtime.extension = extSlick
		}
	}
}

// WithDependencies add dependencies into runtime
func WithDependencies(dependencies ...string) CompileOption {
	return func(runtime *CompilerRuntime) {
		runtime.dependencies = append(runtime.dependencies, dependencies...)
	}
}

// WithAddArguments add external arguments into runtime
func WithAddArguments(arguments ...string) CompileOption {
	return func(runtime *CompilerRuntime) {
		runtime.arguments = append(runtime.arguments, arguments...)
	}
}

// WithSourceRelative sets source_relative when compile
func WithSourceRelative(relative bool) CompileOption {
	return func(runtime *CompilerRuntime) {
		runtime.sourceRelative = relative
	}
}

// WithOutput sets all output directory
func WithOutput(output string) CompileOption {
	return func(runtime *CompilerRuntime) {
		runtime.output = output
	}
}
