package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	// Custom flag handling
	namespace := flag.String("n", "default", "Namespace to query pods from")
	container := flag.String("c", "", "Container name (optional)")
	since := flag.String("since", "", "Only return logs newer than a relative duration")
	limit := flag.Int64("tail", -1, "Number of lines to show from the end of the logs")
	filter := flag.String("filter", "", "Regex to filter pod names (mandatory)")
	help := flag.Bool("help", false, "Show usage information")
	flag.Usage = printUsage

	// Parse only known flags
	flag.Parse()

	// All remaining arguments are for kubectl
	extraArgs := flag.Args()

	if *help {
		printUsage()
		return
	}

	if *filter == "" {
		fmt.Println("Error: --filter argument is mandatory.")
		printUsage()
		os.Exit(1)
	}

	// Compile the regex
	regex, err := regexp.Compile(*filter)
	if err != nil {
		fmt.Printf("Error compiling regex: %v\n", err)
		os.Exit(1)
	}

	// Create Kubernetes client
	clientset, err := createK8sClient()
	if err != nil {
		fmt.Printf("Error creating Kubernetes client: %v\n", err)
		os.Exit(1)
	}

	// Fetch all pods in the namespace
	pods, err := clientset.CoreV1().Pods(*namespace).List(context.TODO(), v1.ListOptions{})
	if err != nil {
		fmt.Printf("Error listing pods: %v\n", err)
		os.Exit(1)
	}

	if len(pods.Items) == 0 {
		fmt.Printf("No pods found in namespace '%s'\n", *namespace)
		return
	}

	// Aggregate logs from matching pods
	matched := 0
	for _, pod := range pods.Items {
		if !regex.MatchString(pod.Name) {
			continue
		}
		matched++
		fmt.Printf("\n--- Logs from pod: %s ---\n", pod.Name)

		// Construct kubectl logs arguments
		cmdArgs := []string{"logs", pod.Name, "-n", *namespace}
		if *container != "" {
			cmdArgs = append(cmdArgs, "-c", *container)
		}
		if *since != "" {
			cmdArgs = append(cmdArgs, "--since", *since)
		}
		if *limit > -1 {
			cmdArgs = append(cmdArgs, "--tail", fmt.Sprintf("%d", *limit))
		}
		cmdArgs = append(cmdArgs, extraArgs...)

		// Execute kubectl logs command
		runKubectlCommand(cmdArgs)
	}

	if matched == 0 {
		fmt.Printf("No pods matched the regex '%s' in namespace '%s'\n", *filter, *namespace)
	}
}

// printUsage displays usage information for the plugin
func printUsage() {
	fmt.Println(`Usage:
kuddle [options] --filter <regex> [additional kubectl logs flags]

Options:
  -n, --namespace <namespace>   Namespace to query pods from (default: "default")
  -c, --container <name>        Container name (optional)
      --since <duration>        Only return logs newer than a relative duration
      --tail <lines>            Number of lines to show from the end of the logs
      --filter <regex>          Regex to filter pod names (mandatory)
  --help                        Show this usage information`)
}

// createK8sClient creates and returns a Kubernetes client
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

// runKubectlCommand runs a kubectl command with the given arguments
func runKubectlCommand(args []string) {
	cmd := exec.Command("kubectl", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("Error executing kubectl command: %v\n", err)
	}
}
