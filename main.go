package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	flag "github.com/spf13/pflag"
)

type (
	Options struct {
		flagFile       string
		startDirectory string
		stopDirectory  string
		help           bool
		quiet          bool
	}
)

var options Options

func init() {
	flag.StringVarP(&options.flagFile, "flag-file", "f", "", "Stop searching if a directory contains this file or directory name")
	flag.StringVarP(&options.stopDirectory, "stop-directory", "s", "", "Stop searching at this directory")
	flag.StringVarP(&options.startDirectory, "start-directory", "d", ".", "Start searching at this directory (default '.')")
	flag.BoolVarP(&options.help, "help", "h", false, "Print this help message")
	flag.BoolVarP(&options.quiet, "quiet", "q", false, "Do not print error messages")
}

func Usage(out io.Writer) {
	fmt.Fprintf(out, "%s: usage: %s [options] filename\n", os.Args[0], os.Args[0])
	fmt.Fprintf(out, "\nOptions:\n\n%s\n", flag.CommandLine.FlagUsages())
}

func ErrUsage() {
	Usage(os.Stderr)
	os.Exit(2)
}

func errorAndExit(msg string, params ...any) {
	if !options.quiet {
    fmt.Fprintf(os.Stderr, "ERROR: ")
		fmt.Fprintf(os.Stderr, msg, params...)
	}

	os.Exit(1)
}

func main() {
	flag.Usage = ErrUsage
	flag.Parse()

	if options.help {
		Usage(os.Stdout)
		os.Exit(0)
	}

	if flag.NArg() != 1 {
		errorAndExit("Missing target file name")
		os.Exit(2)
	}

	targetFile := flag.Arg(0)

	// Start the search from the specified directory
	foundFile, err := searchFile(options.startDirectory, targetFile, options.flagFile, options.stopDirectory)
	if err != nil {
		errorAndExit("%v\n", err)
	}

	fmt.Printf("%s\n", foundFile)
}

func searchFile(startDir, targetFile, flagFile, stopDir string) (string, error) {
	var currentDir string
	var stopDirAbs string
	var err error

	currentDir, err = filepath.Abs(startDir)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path of start directory: %w", err)
	}

	if stopDir != "" {
		stopDirAbs, err = filepath.Abs(stopDir)
		if err != nil {
			return "", fmt.Errorf("failed to get absolute path of stop directory: %w", err)
		}
	}

	for {
		// Check if we've reached the stop directory
		if stopDirAbs != "" {
			if currentDir == stopDirAbs {
				break
			}
		}

		// Check if the target file exists in the current directory
		targetPath := filepath.Join(currentDir, targetFile)
		if _, err := os.Stat(targetPath); err == nil {
			return targetPath, nil
		}

		// Check if the flag file or directory exists in the current directory
		if flagFile != "" {
			flagFilePath := filepath.Join(currentDir, flagFile)
			if _, err := os.Stat(flagFilePath); err == nil {
				break
			}
		}

		// Get the parent directory
		parentDir := filepath.Dir(currentDir)
		if parentDir == currentDir {
			// If the parent directory is the same as the current directory, we've reached the root
			break
		}
		currentDir = parentDir
	}

	return "", fmt.Errorf("file not found")
}
