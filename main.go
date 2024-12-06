package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

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
		fmt.Printf("\n--- Logs from pod: %s ---\n", pod.Name)
		cmdArgs := []string{"logs", pod.Name, "-n", *namespace}
		cmdArgs = append(cmdArgs, extraArgsList...)
		runKubectlCommand(cmdArgs)
	}

	if matched == 0 {
		fmt.Printf("No pods matched the regex '%s' in namespace '%s'\n", *filter, *namespace)
	}
}

func printUsage() {
	fmt.Println(`Usage:
kuddle [options] --filter <regex> [additional kubectl logs flags]

Options:
  --filter <regex>              Regex to filter pod names (mandatory)
  -n, --namespace <namespace>   Namespace to query pods from (default: "default")
  --extraArgs                   flags to be passed into kubectl logs command
  --help                        Show this usage information`)
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

func runKubectlCommand(args []string) {
	fmt.Println(args)

	cmd := exec.Command("kubectl", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("Error executing kubectl command: %v\n", err)
	}
}
