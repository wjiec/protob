package subcommand

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"protob/internal/protob"
	"protob/pkg/logging"
	"protob/pkg/os/fs"
	"protob/pkg/protobuf"
	"protob/pkg/protobuf/gogo"
	"protob/pkg/zip"
	"runtime"
	"strings"

	"github.com/google/go-github/v33/github"
	"github.com/spf13/cobra"
)

var (
	httpClient   = &http.Client{}
	githubClient = github.NewClient(httpClient)
	extensions   = []string{"protoc-gen-gogofast", "protoc-gen-gogofaster", "protoc-gen-gogoslick"}
)

func Install() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install Protobuf compiler and dependencies",
		PreRun: func(cmd *cobra.Command, args []string) {
			if proxy, err := cmd.PersistentFlags().GetString("proxy"); err == nil && proxy != "" {
				httpClient.Transport = &http.Transport{
					Proxy: func(_ *http.Request) (*url.URL, error) {
						return url.Parse(proxy)
					},
				}
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			if err := installProtobuf(cmd.Context()); err == nil {
				_ = installGoGoProtobuf(cmd.Context())
			}
		},
	}

	cmd.PersistentFlags().String("proxy", "", "proxy for http request")

	return cmd
}

// installProtobuf install protobuf compiler and google dependencies
func installProtobuf(ctx context.Context) (err error) {
	logging.Loading("fetch protobuf latest release", func(bar *logging.Bar) {
		defer func() { bar.Error(err) }()

		var release *github.RepositoryRelease
		// https://github.com/protocolbuffers/protobuf
		if release, err = latestRelease(ctx, "protocolbuffers", "protobuf"); err != nil {
			return
		}

		var content []byte
		bar.Text(fmt.Sprintf("downloading protobuf version %s", release.GetTagName()))
		if content, err = downloadRelease(ctx, release); err != nil {
			return
		}

		bar.Text(fmt.Sprintf("extracting resources into %s", protob.Home()))
		if err := extractProtobuf(content, protob.Home()); err != nil {
			bar.Fatal(err.Error())
			return
		}

		bar.Success("protobuf installed")
	})
	return
}

// installGoGoProtobuf install gogo compiler plugins and gogo dependencies
func installGoGoProtobuf(ctx context.Context) (err error) {
	logging.Loading("fetch gogo latest release", func(bar *logging.Bar) {
		defer func() { bar.Error(err) }()

		var release *github.RepositoryRelease
		// https://github.com/gogo/protobuf
		if release, err = latestRelease(ctx, "gogo", "protobuf"); err != nil {
			return
		}

		var content []byte
		bar.Text(fmt.Sprintf("downloading gogo version %s", release.GetTagName()))
		if content, err = downloadContent(ctx, release.GetZipballURL()); err != nil {
			return
		}

		bar.Text(fmt.Sprintf("extracting resources into %s", protob.Temporary()))
		if err = extractGoGoProtobuf(content, protob.Temporary(), protob.Dependency()); err != nil {
			return
		}

		bar.Text(fmt.Sprintf("compiling gogo plugins"))
		if err = compileGoGoExtensions(protob.Temporary(), protob.Home()); err != nil {
			return
		}

		bar.Text(fmt.Sprintf("cleaning temporary directory"))
		if err = os.RemoveAll(protob.Temporary()); err != nil {
			return
		}

		bar.Success("gogo installed")
	})
	return
}

// latestRelease retrieve latest release info from github
func latestRelease(ctx context.Context, owner, repo string) (*github.RepositoryRelease, error) {
	releases, _, err := githubClient.Repositories.ListReleases(ctx, owner, repo, &github.ListOptions{PerPage: 1})
	if err != nil {
		return nil, err
	} else if len(releases) == 0 {
		return nil, errors.New("download: release not found")
	}

	return releases[0], nil
}

// downloadRelease download compiler asset matched system
func downloadRelease(ctx context.Context, release *github.RepositoryRelease) ([]byte, error) {
	for _, asset := range release.Assets {
		switch runtime.GOOS {
		case "windows":
			if strings.Contains(asset.GetName(), "win64") {
				return downloadContent(ctx, asset.GetBrowserDownloadURL())
			}
		case "linux":
			if strings.Contains(asset.GetName(), "linux-x86_64") {
				return downloadContent(ctx, asset.GetBrowserDownloadURL())
			}
		case "darwin":
			if strings.Contains(asset.GetName(), "osx-x86_64") {
				return downloadContent(ctx, asset.GetBrowserDownloadURL())
			}
		}
	}

	return nil, errors.New("install: unable to match assert")
}

// downloadContent download url into bytes buffer
func downloadContent(ctx context.Context, url string) ([]byte, error) {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer func() { _ = resp.Body.Close() }()
	return ioutil.ReadAll(resp.Body)
}

// extractProtobuf extract compiler/dependencies from zipball content into dir
func extractProtobuf(content []byte, dir string) error {
	return zip.VisitFiles(content, func(file *zip.File) error {
		if strings.HasPrefix(file.Name, "bin/") {
			return zip.AsReader(file, func(reader io.Reader) error {
				return fs.WriteFile(fs.Join(dir, protobuf.CompilerExecutable), reader, fs.ExecutableFilePerm)
			})
		} else if strings.HasPrefix(file.Name, "include/") {
			return zip.AsReader(file, func(reader io.Reader) error {
				return fs.WriteFile(fs.Join(dir, file.Name), reader, fs.RegularFilePerm)
			})
		}

		return nil
	})
}

// extractGoGoProtobuf extract gogo sources from zipball content into temp and include
func extractGoGoProtobuf(content []byte, temp string, include string) error {
	return zip.VisitFiles(content, func(file *zip.File) error {
		return zip.AsReader(file, func(reader io.Reader) error {
			filename := fs.Children(file.Name)
			if strings.HasPrefix(filename, "gogoproto") && strings.HasSuffix(filename, ".proto") {
				if err := fs.WriteFile(fs.Join(include, gogo.Namespace, filename), reader, fs.RegularFilePerm); err != nil {
					return err
				}
			}
			return fs.WriteFile(fs.Join(temp, filename), reader, fs.RegularFilePerm)
		})
	})
}

// compileGoGoExtensions compile protoc-gen-gogo* extensions from source into dst
func compileGoGoExtensions(source string, dst string) (err error) {
	var compiler string
	if compiler, err = exec.LookPath("go"); err != nil {
		return errors.New("install: go compiler not found")
	}

	for _, extension := range extensions {
		binary, input := fs.Join(dst, extension), fs.Join(source, extension, "main.go")
		if runtime.GOOS == "windows" {
			binary += ".exe"
		}

		if _, err = exec.Command(compiler, "build", "-o", binary, input).Output(); err != nil {
			return err
		}
	}

	return nil
}
