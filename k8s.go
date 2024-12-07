package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func CreateK8sClient() (*kubernetes.Clientset, error) {
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

func ListPods(namespace string) corev1.PodList {
	clientset, err := CreateK8sClient()
	if err != nil {
		fmt.Printf("Error creating Kubernetes client: %v\n", err)
		os.Exit(1)
	}

	pods, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), v1.ListOptions{})
	if err != nil {
		fmt.Printf("Error listing pods: %v\n", err)
		os.Exit(1)
	}

	return *pods
}

func RunKubectlCommandAsync(pod string, args []string) {
	go func() {
		defer wg.Done()

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
