// Command bundle validates Partite skills and builds their release archives.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
)

func run(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("bundle", flag.ContinueOnError)
	flags.SetOutput(stderr)
	check := flags.Bool("check", false, "validate without creating bundles")
	version := flags.String("version", "", "require this release version (without a v prefix)")
	output := flags.String("output", "", "output directory (default: dist under root)")
	root := flags.String("root", ".", "package root (default: current directory)")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}
	if flags.NArg() != 0 {
		fmt.Fprintln(stderr, "Unexpected positional arguments:", flags.Args())
		return 2
	}
	var index manifest
	var err error
	if *check {
		index, err = validate(*root, *version)
	} else {
		index, err = build(*root, *output, *version)
	}
	if err != nil {
		fmt.Fprintln(stderr, "Validation failed:", err)
		return 1
	}
	if *check {
		fmt.Fprintf(stdout, "Validated %d skills and package manifests for %s\n", len(index.Skills), index.Version)
	} else {
		fmt.Fprintf(stdout, "Built %d bundles for %s\n", len(index.Artifacts), index.Version)
		for _, artifact := range index.Artifacts {
			fmt.Fprintf(stdout, "  %s (%d bytes)\n", artifact.File, artifact.Bytes)
		}
	}
	return 0
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}
