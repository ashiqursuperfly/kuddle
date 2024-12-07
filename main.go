package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const VERSION = "0.0.1"

var wg sync.WaitGroup

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

func createK8sClient() (*kubernetes.Clientset, error) {
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		kubeconfig = os.ExpandEnv("$HOME/.kube/config")
	}
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(config)
}

func runKubectlCommandAsync(pod string, args []string) {
	go func() {
		defer wg.Done() // Decrement the WaitGroup counter when the goroutine completes

		cmd := exec.Command("kubectl", args...)
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			WriteKubectlLogs(pod, fmt.Sprintf("Error creating stdout pipe: %v", err))
			return
		}

		stderr, err := cmd.StderrPipe()
		if err != nil {
			WriteKubectlLogs(pod, fmt.Sprintf("Error creating stderr pipe: %v", err))
			return
		}

		if err := cmd.Start(); err != nil {
			WriteKubectlLogs(pod, fmt.Sprintf("Error starting kubectl command: %v", err))
			return
		}

		go func() {
			scanner := bufio.NewScanner(stdout)
			for scanner.Scan() {
				WriteKubectlLogs(pod, scanner.Text())
			}
			if err := scanner.Err(); err != nil {
				WriteKubectlLogs(pod, fmt.Sprintf("Error reading stdout: %v", err))
			}
		}()

		go func() {
			scanner := bufio.NewScanner(stderr)
			for scanner.Scan() {
				WriteKubectlLogs(pod, scanner.Text())
			}
			if err := scanner.Err(); err != nil {
				WriteKubectlLogs(pod, fmt.Sprintf("Error reading stderr: %v", err))
			}
		}()

		if err := cmd.Wait(); err != nil {
			WriteKubectlLogs(pod, fmt.Sprintf("Error waiting for kubectl command: %v", err))
		}
	}()
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

	clientset, err := createK8sClient()
	if err != nil {
		fmt.Printf("Error creating Kubernetes client: %v\n", err)
		os.Exit(1)
	}

	pods, err := clientset.CoreV1().Pods(*namespace).List(context.TODO(), v1.ListOptions{})
	if err != nil {
		fmt.Printf("Error listing pods: %v\n", err)
		os.Exit(1)
	}

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

		// Increment the WaitGroup counter
		wg.Add(1)

		// Run the command in a new goroutine
		runKubectlCommandAsync(pod.Name, cmdArgs)
	}

	// Wait for all goroutines to finish
	wg.Wait()

	fmt.Println("Finished showing logs from %s pods", matched)
	if matched == 0 {
		fmt.Printf("No pods matched the regex '%s' in namespace '%s'\n", *filter, *namespace)
	}
}

