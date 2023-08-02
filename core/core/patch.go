package core

import (
	"encoding/json"
	"errors"
	"os"

	logger "github.com/kubescape/go-logger"

	"context"
	"strings"

	ksmetav1 "github.com/kubescape/kubescape/v2/core/meta/datastructures/v1"

	"github.com/kubescape/storage/pkg/apis/softwarecomposition/v1beta1"
	"github.com/kubescape/storage/pkg/generated/clientset/versioned"
	copa "github.com/project-copacetic/copacetic/pkg/patch"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
)

func (ks *Kubescape) Patch(ctx context.Context, patchInfo *ksmetav1.PatchInfo) error {

	logger.L().Info("Finding the image vulnerability report file...")

	kubeconfig := patchInfo.Kubeconfig

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return err
	}

	client, err := versioned.NewForConfig(config)
	if err != nil {
		return err
	}

	vulnerabilityManifests, err := client.SpdxV1beta1().VulnerabilityManifests("kubescape").List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}

	// Replace "/" and ":" with "-" in image tag
	imageTag := strings.ReplaceAll(patchInfo.ImageTag, "/", "-")
	imageTag = strings.ReplaceAll(imageTag, ":", "-")

	vulnManifest := v1beta1.VulnerabilityManifest{}

	for _, vulnerabilityManifest := range vulnerabilityManifests.Items {
		if strings.Contains(vulnerabilityManifest.Name, imageTag) {
			logger.L().Info("Found the image vulnerability report file: " + vulnerabilityManifest.Name)
			getVulnManifest, err := client.SpdxV1beta1().VulnerabilityManifests("kubescape").Get(ctx, vulnerabilityManifest.Name, metav1.GetOptions{})
			if err != nil {
				return err
			}
			vulnManifest = *getVulnManifest
		}
	}

	// Check if vulnManifest is empty
	if vulnManifest.Name == "" {
		return errors.New("Could not find the image vulnerability report file for the image: " + patchInfo.ImageTag)
	}

	// TODO: Improve the file saving logic by reading line by line, and not loading the entire file into memory
	// Save the found Vulnerability Manifest to a file in JSON format
	data, err := json.MarshalIndent(vulnManifest, "", "\t")
	if err != nil {
		return err
	}
	path := "./" + vulnManifest.Name + ".json"
	if err := os.WriteFile(path, data, 0644); err != nil {
		return err
	}

	// Patch images with copa
	if err := copa.Patch(ctx, patchInfo.Timeout, patchInfo.BuildkitAddress, patchInfo.ImageTag, path, patchInfo.PatchedImageTag, ""); err != nil {
		return err
	}

	// Remove the image vulnerability report file after patching
	if err := os.Remove(path); err != nil {
		return err
	}

	logger.L().Info("Successfully patched the image. Run the patched image in your cluster to find the new vulnerabilities.")

	return nil
}
