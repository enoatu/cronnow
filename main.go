package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/ktr0731/go-fuzzyfinder"
)

func usage() {
	fmt.Println("Usage: cronexec [-f file] [-h]")
	fmt.Println("  -f file   Specify a crontab file. If not specified, the output of 'crontab -l' is used.")
	fmt.Println("  -h        Display this help message.")
}

func main() {
	filePtr := flag.String("f", "", "Specify a crontab file. If not specified, the output of 'crontab -l' is used.")
	helpPtr := flag.Bool("h", false, "Display this help message.")
	flag.Parse()

	if *helpPtr {
		usage()
		return
	}

	var cronContent string
	if *filePtr != "" {
		if _, err := os.Stat(*filePtr); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "The specified file '%s' does not exist.\n", *filePtr)
			os.Exit(1)
		}
		data, err := os.ReadFile(*filePtr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to read the file: %v\n", err)
			os.Exit(1)
		}
		cronContent = string(data)
	} else {
		out, err := exec.Command("crontab", "-l").Output()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to execute 'crontab -l': %v\n", err)
			os.Exit(1)
		}
		cronContent = string(out)
	}

	// Split content into lines
	lines := strings.Split(cronContent, "\n")
	// Regular expression to match environment variable definitions
	envRegex := regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*=`)

	var envLines []string
	var jobLines []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		// Skip comment lines
		if strings.HasPrefix(trimmed, "#") {
			continue
		}
		// If it's an environment variable definition, add to envLines;
		// otherwise, treat it as a cron job line.
		if envRegex.MatchString(trimmed) {
			envLines = append(envLines, trimmed)
		} else {
			jobLines = append(jobLines, trimmed)
		}
	}

	// Set environment variables from the crontab definitions
	for _, envLine := range envLines {
		parts := strings.SplitN(envLine, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			os.Setenv(key, value)
		}
	}

	if len(jobLines) == 0 {
		fmt.Println("No cron jobs found.")
		os.Exit(1)
	}

	// Use go-fuzzyfinder to interactively select a job line
	idx, err := fuzzyfinder.Find(jobLines, func(i int) string {
		return jobLines[i]
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to select a job: %v\n", err)
		os.Exit(1)
	}

	selected := jobLines[idx]

	// Split the selected line by spaces and remove the first 5 fields (schedule)
	fields := strings.Fields(selected)
	if len(fields) < 6 {
		fmt.Println("The selected line is not in a valid cron job format (insufficient fields).")
		os.Exit(1)
	}
	cmdLine := strings.Join(fields[5:], " ")
	// Expand environment variables in the command
	expandedCmd := os.ExpandEnv(cmdLine)

	fmt.Printf("Command to be executed: %s\n", expandedCmd)
	fmt.Print("Are you sure you want to execute? (y/n): ")

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	answer := strings.TrimSpace(scanner.Text())
	if answer != "y" && answer != "Y" {
		fmt.Println("Operation cancelled.")
		os.Exit(0)
	}

	// Execute the command using bash
	execCmd := exec.Command("bash", "-c", expandedCmd)
	execCmd.Stdout = os.Stdout
	execCmd.Stderr = os.Stderr
	execCmd.Stdin = os.Stdin
	if err := execCmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to execute the command: %v\n", err)
		os.Exit(1)
	}
}
