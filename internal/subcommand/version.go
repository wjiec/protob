package subcommand

import (
	"os"
	"protob/internal/protob"
	"protob/pkg/protobuf"
	"runtime"
	"text/template"

	"github.com/spf13/cobra"
)

const versionTemplate = `CommandLine: protob
 Version: {{.Version}}
 Go Version: {{.GoVersion}}
 Built: {{.BuildTime}}(Rev: {{.GitRevision}})
 Base Dir: {{.BaseDir}}

Protobuf: protoc
{{- if .SysCompilerVersion }}
 System Compiler: {{ .SysCompilerVersion }}(system)
{{- end }}
{{- if .EmbeddedCompilerVersion }}
 Embedded Compiler: {{ .EmbeddedCompilerVersion }}(embedded)
{{- end }}
 Include Path: {{ .IncludePath }}
{{- if and (not .SysCompilerVersion) (not .EmbeddedCompilerVersion) }}

Please run 'protob install' to install protobuf compiler
{{- end}}
`

func Version(version, revision, buildTime string) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version info",
		Run: func(cmd *cobra.Command, args []string) {
			data := map[string]string{
				"Version":     version,
				"GoVersion":   runtime.Version(),
				"BuildTime":   buildTime,
				"GitRevision": revision,
				"BaseDir":     protob.Home(),
				"IncludePath": protob.Dependency(),
			}

			if compiler, err := protobuf.NewSystemCompiler(); err == nil {
				data["SysCompilerVersion"] = compiler.Version
			}
			if compiler, err := protobuf.NewCompiler(protob.Compiler()); err == nil {
				data["EmbeddedCompilerVersion"] = compiler.Version
			}

			tpl, _ := template.New("version").Parse(versionTemplate)
			_ = tpl.Execute(os.Stdout, data)
		},
	}
}
