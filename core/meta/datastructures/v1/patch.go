package v1

import "time"

type PatchInfo struct {
	ImageTag        string        // image tag to be patched
	PatchedImageTag string        // can be empty, if empty then the image tag will be patched with the latest tag
	BuildkitAddress string        // buildkit address
	Kubeconfig      string        // kubeconfig file
	Timeout         time.Duration // timeout for patching an image
}
