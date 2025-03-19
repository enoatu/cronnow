# cronnow

**cronnow** is an interactive command-line tool that brings your cron jobs to life on demand. With an intuitive fuzzy finder interface, you can effortlessly browse, select, and execute cron jobs directly from your crontab, making job testing and debugging a breeze.

## Features

- **Interactive Selection:** Quickly find and select the desired cron job using a smart fuzzy finder.
- **Flexible Cron Input:** Automatically reads your current crontab or a user-specified file.
- **Environment Variable Handling:** Automatically sets and expands environment variables defined in your crontab.
- **Safe Execution:** Review the command before execution or skip confirmation with the `-y` flag for non-interactive usage.
- **Customizable Shell Execution:** By default, executes commands using bash; you can easily adapt it to your shell of choice.

## Installation

Ensure you have [Go](https://golang.org/) installed. Then, you can install **cronnow** via:

```sh
go get github.com/enoatu/cronnow
```

Or clone the repository and build manually:

```sh
git clone https://github.com/enoatu/cronnow.git
cd cronnow
go build -o cronnow
```

or use [mise en place](https://mise.jdx.dev/) to install the tool:

```sh
mise settings experimental=true
mise use -g go go:github.com/enoatu/cronnow
```

Note: The above command will install go if it's not already installed.

## Usage

Launch **cronnow** to interactively select and execute cron jobs:

```sh
cronnow
```

### Options

- **`-f file`**  
  Specify a crontab file instead of using the output of `crontab -l`.

- **`-y`**  
  Automatically execute the selected cron job without prompting for confirmation.

- **`-h`**  
  Display the help message.

For example, to run a job from a specific file without confirmation:

```sh
cronnow -f path/to/crontab -y
```

## How It Works

1. **Parsing the Crontab:**  
   The tool reads your crontab, separating environment variable definitions from cron job entries.

2. **Fuzzy Finder Interface:**  
   It then leverages a fuzzy finder to allow you to easily pick a job from the list.

3. **Command Preparation:**  
   After selecting a job, **cronnow** expands any environment variables in the command line and prepares it for execution.

4. **Execution:**  
   The command is executed in bash (or your shell of choice), giving you a familiar execution environment.

## Contributing

Contributions, issues, and feature requests are very welcome! Check out the [issues page](https://github.com/enoatu/cronnow/issues) if you have any questions or ideas.

## License

This project is licensed under the MIT License.

---

Experience your cron jobs like never before—execute them on your terms, whenever you need them, with **cronnow**!

Feel free to modify and enhance this README to best suit your project's style and vision. Enjoy coding!
