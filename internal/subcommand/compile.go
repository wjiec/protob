package subcommand

import (
	"protob/internal/protob"
	"protob/pkg/logging"
	"protob/pkg/protobuf"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func Compile() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "compile",
		Short: "Compile Protobuf files",
		Run: func(cmd *cobra.Command, args []string) {
			compiler, err := protobuf.NewCompiler(protob.Compiler())
			if sys, _ := cmd.PersistentFlags().GetBool("sys"); sys || err != nil {
				compiler, _ = protobuf.NewSystemCompiler()
			}

			if compiler == nil {
				logging.Fatal("compile: compiler not found or invalid")
				return
			}

			runtime, targets := buildRuntimeAndTarget(cmd.PersistentFlags(), args)
			for _, target := range targets {
				if err := compiler.Compile(target, runtime); err != nil {
					logging.Fatal("compile: build '%s' error: %s", target, err)
				}
			}
			logging.Success("build completed")
		},
	}

	cmd.PersistentFlags().Bool("sys", false, "using system compiler")
	cmd.PersistentFlags().StringP("output", "o", "", "output directory")

	cmd.PersistentFlags().Bool("fast", false, "enable gogo-fast extension")
	cmd.PersistentFlags().Bool("faster", false, "enable gogo-faster extension")
	cmd.PersistentFlags().Bool("slick", true, "enable gogo-slick extension")
	cmd.PersistentFlags().Bool("grpc", false, "whether compile with grpc")

	cmd.PersistentFlags().StringSliceP("proto_path", "I", nil, "transparent argument for protoc set dependencies")
	cmd.PersistentFlags().Bool("source-relative", false, "transparent argument for protoc set source_relative")

	return cmd
}

// buildRuntimeAndTarget build compile runtime and split targets
func buildRuntimeAndTarget(fs *pflag.FlagSet, args []string) (*protobuf.CompilerRuntime, []string) {
	var options []protobuf.CompileOption
	if grpc, err := fs.GetBool("grpc"); err == nil {
		options = append(options, protobuf.WithGrpc(grpc))
	}
	if fast, err := fs.GetBool("fast"); err == nil {
		options = append(options, protobuf.WithExtFast(fast))
	}
	if faster, err := fs.GetBool("faster"); err == nil {
		options = append(options, protobuf.WithExtFaster(faster))
	}
	if slick, err := fs.GetBool("slick"); err == nil {
		options = append(options, protobuf.WithExtSlick(slick))
	}
	if deps, err := fs.GetStringSlice("proto_path"); err == nil && deps != nil {
		options = append(options, protobuf.WithDependencies(deps...))
	}
	if relative, err := fs.GetBool("source_relative"); err == nil {
		options = append(options, protobuf.WithSourceRelative(relative))
	}
	if output, err := fs.GetString("output"); err == nil && output != "" {
		options = append(options, protobuf.WithOutput(output))
	}

	var targets []string
	for _, arg := range args {
		if arg[0] == '-' {
			options = append(options, protobuf.WithAddArguments(arg))
		} else {
			targets = append(targets, arg)
		}
	}

	return protobuf.NewCompileRuntime(options...), targets
}
