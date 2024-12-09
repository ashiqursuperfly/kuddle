package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"
)

const VERSION = "0.0.1"

func printUsage() {
	fmt.Printf(`Usage:
kuddle [options] --filter <regex> [additional kubectl logs flags]
%s
Options:
  --filter <regex>              Regex to filter pod names (mandatory). Must be matchable using go regexp: regex.MatchString
  -n, --namespace <namespace>   Namespace to query pods from (default: "default")
  --extraArgs                   flags to be passed into kubectl logs command
  --help                        Show this usage information`, VERSION)
}

func main() {
	namespace := flag.String("n", "default", "Namespace to query pods from")
	filter := flag.String("filter", "", "Regex to filter pod names (mandatory)")
	extraArgs := flag.String("extraArgs", "", "Extra args to be passed on to kubectl")
	help := flag.Bool("help", false, "Show usage information")
	flag.Usage = printUsage

	flag.Parse()

	extraArgsList := strings.Split(*extraArgs, " ")

	if *help {
		printUsage()
		return
	}

	if *filter == "" {
		fmt.Println("Error: --filter argument is mandatory.")
		printUsage()
		os.Exit(1)
	}

	regex, err := regexp.Compile(*filter)
	if err != nil {
		fmt.Printf("Error compiling regex: %v\n", err)
		os.Exit(1)
	}

	pods := ListPods(*namespace)

	if len(pods.Items) == 0 {
		fmt.Printf("No pods found in namespace '%s'\n", *namespace)
		return
	}

	matched := 0
	for _, pod := range pods.Items {
		if !regex.MatchString(pod.Name) {
			continue
		}
		matched++
		cmdArgs := []string{"logs", pod.Name, "-n", *namespace}
		cmdArgs = append(cmdArgs, extraArgsList...)

		// Increment the WaitGroup counter and run the command in a new goroutine
		wg.Add(1)
		RunKubectlCommandAsync(pod.Name, cmdArgs)
	}

	// Wait for all goroutines to finish
	wg.Wait()

	fmt.Printf("Finished showing logs from %v pods\n", matched)
	if matched == 0 {
		fmt.Printf("No pods matched the regex '%s' in namespace '%s'\n", *filter, *namespace)
	}
}
