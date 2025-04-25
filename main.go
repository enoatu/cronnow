package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/exec"
    "runtime"
	"regexp"
	"strings"

    "github.com/koki-develop/go-fzf"
)

func usage() {
	fmt.Println("Usage: cronnow [-f file] [-h]")
	fmt.Println("  -f file   Specify a crontab file. If not specified, the output of 'crontab -l' is used.")
	fmt.Println("  -y        Execute the selected cron job without confirmation.")
	fmt.Println("  -h        Display this help message.")
    fmt.Println("  -d        Debug mode.")
}

type CronEnv struct {
	Shell string
	Path  string
}

func detectOS() string {
	goos := runtime.GOOS

	if goos == "darwin" {
		return "macos"
	}
	if goos != "linux" {
		return goos // fallback: windows, etc.
	}

	file, err := os.Open("/etc/os-release")
	if err != nil {
		return "linux"
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "ID=") {
			return strings.Trim(strings.SplitN(line, "=", 2)[1], `"`)
		}
	}
	return "linux"
}

func getCronEnv(id string) CronEnv {
	switch id {
    case "ubuntu":
			return CronEnv{Shell: "/bin/sh", Path: "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/usr/games:/usr/local/games:/snap/bin"}
    case "debian":
			return CronEnv{Shell: "/bin/sh", Path: "/usr/bin:/bin"}
		case "alpine":
			return CronEnv{Shell: "/bin/sh", Path: "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"}
		case "rocky":
			return CronEnv{Shell: "/bin/bash", Path: "/sbin:/bin:/usr/sbin:/usr/bin"}
		case "macos":
			return CronEnv{Shell: "/bin/sh", Path: "/usr/bin:/bin"}
		case "arch":
			return CronEnv{Shell: "/bin/sh", Path: "/usr/bin:/bin"}
		default:
			return CronEnv{Shell: "/bin/sh", Path: "/usr/bin:/bin"}
	}
}

func main() {
	filePtr := flag.String("f", "", "Specify a crontab file. If not specified, the output of 'crontab -l' is used.")
    autoConfirmPtr := flag.Bool("y", false, "Execute the selected cron job without confirmation.")
	helpPtr := flag.Bool("h", false, "Display this help message.")
    debugPtr := flag.Bool("d", false, "Debug mode.")
	flag.Parse()

	if *helpPtr {
		usage()
		return
	}

    // OSごとのcron環境変数を設定
	id := detectOS()
    if *debugPtr {
        fmt.Fprintf(os.Stderr, "Detected OS: %s\n", id)
    }
	cronEnv := getCronEnv(id)

	os.Setenv("SHELL", cronEnv.Shell)
	os.Setenv("PATH", cronEnv.Path)

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

	lines := strings.Split(cronContent, "\n")
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

    f, _ := fzf.New()

	idxs, err := f.Find(jobLines, func(i int) string {
		return jobLines[i]
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to select a job: %v\n", err)
		os.Exit(1)
	}

    idx := idxs[0]

	selected := jobLines[idx]

	fields := strings.Fields(selected)
	if len(fields) < 6 {
		fmt.Println("The selected line is not in a valid cron job format (insufficient fields).")
		os.Exit(1)
	}
	cmdLine := strings.Join(fields[5:], " ")
	expandedCmd := os.ExpandEnv(cmdLine)

	fmt.Printf("Command to be executed: %s\n\n", expandedCmd)

	if !*autoConfirmPtr {
		fmt.Print("Are you sure you want to execute? (y/n): ")
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		answer := strings.TrimSpace(scanner.Text())
		if answer != "y" && answer != "Y" {
			fmt.Println("Operation cancelled.")
			os.Exit(0)
		}
	}

	execCmd := exec.Command(expandedCmd)
  execCmd.Stdout = os.Stdout
  execCmd.Stderr = os.Stderr
  execCmd.Stdin = os.Stdin
	if err := execCmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to execute the command: %v\n", err)
		os.Exit(1)
	}
}
