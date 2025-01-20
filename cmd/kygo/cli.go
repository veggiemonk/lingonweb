package main

import (
	"flag"
	"fmt"
	"os"
	"runtime/debug"
	"strings"

	"github.com/veggiemonk/lingonweb/knowntypes"
	"github.com/golingon/lingon/pkg/kube"
	"github.com/golingon/lingon/pkg/kubeutil"
	"golang.org/x/exp/slog"
)

const crdMsg = "IF there is an issue with CRDs. Please visit this page to solve it https://github.com/golingon/lingon/tree/main/docs/kubernetes/crd"

func main() {
	var in, out, appName, pkgName string
	var version, verbose, ignoreErr bool
	var groupByKind, removeAppName bool

	fs := flag.NewFlagSet("kygo", flag.ExitOnError)
	fs.SetOutput(os.Stdout)

	fs.StringVar(
		&in,
		"in",
		"-",
		"specify the input directory of the yaml manifests, '-' for stdin",
	)
	fs.StringVar(
		&out,
		"out",
		"out",
		"specify the output directory for manifests.",
	)
	fs.StringVar(
		&appName,
		"app",
		"myapp",
		"specify the app name. This will be used as the package name if none is specified.",
	)
	fs.StringVar(
		&pkgName,
		"pkg",
		"lingon",
		"specify the package name. If none is specified the app name will be used. Cannot contain a dash.",
	)
	fs.BoolVar(
		&groupByKind,
		"group",
		true,
		"specify if the output should be grouped by kind (default) or split by name.",
	)
	fs.BoolVar(
		&removeAppName,
		"clean-name",
		true,
		"specify if the app name should be removed from the variable, struct and file name.",
	)
	fs.BoolVar(
		&ignoreErr,
		"ignore-errors",
		false,
		"ignore errors, useful to generate as much as possible",
	)

	fs.BoolVar(&verbose, "v", false, "show verbose logs")
	fs.BoolVar(&version, "version", false, "show version")

	if err := fs.Parse(os.Args[1:]); err != nil {
		panic(err)
	}

	if version {
		printVersion()
		return
	}

	pkgName = strings.ReplaceAll(pkgName, "-", "_")

	slog.Info(
		"flags",
		slog.String("in", in),
		slog.String("out", out),
		slog.String("app", appName),
		slog.Bool("group", groupByKind),
		slog.Bool("clean-name", removeAppName),
		slog.Bool("verbose", verbose),
		slog.Bool("ignore-errors", ignoreErr),
	)

	if err := run(
		in,
		out,
		appName,
		pkgName,
		groupByKind,
		removeAppName,
		verbose,
		ignoreErr,
	); err != nil {
		slog.Error(
			"run",
			slog.Any("error", err),
			slog.String("CRD", crdMsg),
		)
		os.Exit(1)
	}

	slog.Info("done")
}

func run(
	in, out, appName, pkgName string,
	groupByKind, removeAppName, verbose, ignoreErr bool,
) error {
	opts := []kube.ImportOption{
		kube.WithImportAppName(appName),
		kube.WithImportPackageName(pkgName),
		kube.WithImportOutputDirectory(out),
		//
		// init function to register types to runtime.NewScheme() in another file
		//
		kube.WithImportSerializer(knowntypes.Codecs.UniversalDeserializer()),
	}
	opts = append(opts, kube.WithImportGroupByKind(groupByKind))
	opts = append(opts, kube.WithImportRemoveAppName(removeAppName))
	opts = append(opts, kube.WithImportVerbose(verbose))
	opts = append(opts, kube.WithImportIgnoreErrors(ignoreErr))

	// stdin
	if in == "-" {
		opts = append(opts, kube.WithImportReadStdIn())
		if err := kube.Import(opts...); err != nil {
			return fmt.Errorf("import: %w", err)
		}
		return nil
	}

	// single file
	fi, err := os.Stat(in)
	if err != nil {
		return err
	}
	if !fi.IsDir() {
		opts = append(opts, kube.WithImportManifestFiles([]string{in}))
		if err := kube.Import(opts...); err != nil {
			return fmt.Errorf("import: %w", err)
		}
		return nil
	}

	// directory
	files, err := kubeutil.ListYAMLFiles(in)
	if err != nil {
		slog.Error("list yaml files", err)
	}

	fmt.Printf("files:\n- %s\n", strings.Join(files, "\n- "))
	opts = append(opts, kube.WithImportManifestFiles(files))
	if err := kube.Import(opts...); err != nil {
		return fmt.Errorf("import: %w", err)
	}
	return nil
}

var (
	ver    = "dev"
	commit = "none"
	date   = "unknown"
)

func printVersion() {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		_, _ = fmt.Fprintln(os.Stderr, "error reading build-info")
		os.Exit(1)
	}
	fmt.Printf("Build:\n%s\n", bi)
	fmt.Printf("Version: %s\n", ver)
	fmt.Printf("Commit: %s\n", commit)
	fmt.Printf("Date: %s\n", date)
}
