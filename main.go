package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"slices"
	"os/exec"
	"path/filepath"
)

func customParser(input string) []string {
	var result []string
	var current string
	isSingleQuoted := false
	isDoubleQuoted := false

	input = strings.Trim(input, "\r\n")

	for i := 0; i < len(input); i++ {
		c := input[i]
		//handling backslashes in single and double quotes
		if c == '\\' && !isSingleQuoted && !isDoubleQuoted {
			if i+1 < len(input) {
				i++
				current += string(input[i])
			}
		} else if c == '\\' && isDoubleQuoted {
			if i+1 < len(input) && (input[i+1] == '$' || input[i+1] == '\\' || input[i+1] == '"') {
				i++
				current += string(input[i])
			} else {
				current += "\\"
			}
		} else if c == '\'' && !isDoubleQuoted { //support for single quotes
			isSingleQuoted = !isSingleQuoted
		} else if c == '"' && !isSingleQuoted { //support for double quotes
			isDoubleQuoted = !isDoubleQuoted
		} else if c == ' ' && !isSingleQuoted && !isDoubleQuoted {//splitting on space
			if current != "" {
				result = append(result, current)
				current = ""
			}
		} else {
			current += string(c)
		}
	}

	if current != "" {
		result = append(result, current)
	}

	return result
}


func main() {

	// defining shell builtins
	shellBuiltins := []string{"echo", "exit", "type","pwd"}

	//defining valid commands
	validCommands := []string{"cat","ls","cp","mv","grep","rm","mkdir","rmdir","cd","chmod","chown","ps","kill","top","df","du","free","uname","date","who","w","uptime","history","clear","touch","head","tail","sort","uniq","wc","cut","tr","sed","awk","find","tar","gzip","gunzip","zip","unzip","ssh","scp","rsync","curl","wget","ping","traceroute","netstat","ifconfig","route","iptables","tcpdump","dig","host","nslookup","whois","lsof","ps","kill","top","df","du","free","uname","date","who","w","uptime","history","clear","touch","head","tail","sort","uniq","wc","cut","tr","sed","awk","find","tar","gzip","gunzip","zip","unzip","ssh","scp","rsync","curl","wget","ping","traceroute","netstat","ifconfig","route","iptables","tcpdump","dig","host","nslookup","whois","lsof"}

	// REPL loop
	for {

		// Print the shell prompt
		fmt.Fprint(os.Stdout, "$ ")

		// Wait for user input, print error if there is error reading input
		command, err := bufio.NewReader(os.Stdin).ReadString('\n')

		if err != nil {
			fmt.Fprintln(os.Stderr, "Error reading input:", err)
			os.Exit(1)
		}

		// Remove the newline character from the end of the input
		command = strings.TrimSpace(command)	
				
		// Exit the shell if the command is "exit"
		if command == "exit" {
			os.Exit(0)
		}
		
		// Split the command into words
		words := customParser(command)
		if len(words) == 0 {
			continue // Skip empty input
		}
		
		// echo command
		if words[0] == "echo" {
			//support for ''
			if len(words) < 2 {
				continue
			}
			// Join the arguments into a single string
			input := strings.Join(words[1:], " ")

			// Check if the input is enclosed in single quotes
			if strings.HasPrefix(input, "'") && strings.HasSuffix(input, "'") {
				// Remove the enclosing single quotes
				input = input[1 : len(input)-1]
			}

			// Print the rest of the words as the output of the echo command
			fmt.Println(input)
			continue
		}

		// type command
		if words[0] == "type" {
			cmd := words[1]	

			if slices.Contains(shellBuiltins, cmd) {
				fmt.Println(cmd + " is a shell builtin")
				continue
			}
			//extended to find PATH of command
			if !slices.Contains(validCommands, cmd) && !slices.Contains(shellBuiltins, cmd) {
				fmt.Println(cmd + ": not found")
			}else {
				path, err := exec.LookPath(cmd)
				if err != nil {
					fmt.Println(cmd + ": not found")
				} else {
					fmt.Println(cmd + " is " + path)
				}
			}
			continue
		}

		//cat command
		if words[0] == "cat" {
			if len(words) < 2 {
				continue
			}
			var output strings.Builder

			for _, path := range words[1:] {
				file, err := os.Open(path)
				if err != nil {
					fmt.Println("cat: " + path + ": No such file or directory")
					continue
				}
				defer file.Close()

				scanner := bufio.NewScanner(file)
				for scanner.Scan() {
					output.WriteString(scanner.Text())
				}
				if err := scanner.Err(); err != nil {
					fmt.Println("Error reading file:", err)
				}
			}

			// Print all concatenated content in one line
			fmt.Println(strings.TrimSpace(output.String()))
			continue
		}

		//pwd command
		if words[0] == "pwd" {
			dir, err := os.Getwd()
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println(dir)
			continue
		}

		//cd command
		if words[0] == "cd" {
			if len(words) < 2 {
				continue
			}
			//support for .. and . and ~ in cd command
			path := words[1]
			if path == "~" {
				path = os.Getenv("HOME")
			}
			isAbsolute := filepath.IsAbs(path)
			if !isAbsolute {
				cwd, err := os.Getwd()
				if err != nil {
					fmt.Println("Error getting current directory:", err)
					continue
				}
				path = filepath.Join(cwd, path)
			}
			err := os.Chdir(path)
			if err != nil {
				fmt.Println("cd: " + path + ": No such file or directory")
			}	
			continue
		}

		//external command support on any executable command in the PATH
		path, _ := exec.LookPath(words[0])
		if path != "" {
			cmd := exec.Command(path, words[1:]...)
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			err := cmd.Run()
			if err != nil {
				fmt.Println(err)
			}
			continue
		}
		
		fmt.Println(command + ": command not found")
	}
		
}
