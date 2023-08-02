# Patch Command

The patch command is used for patching container images which are running in your cluster with vulnerabilities.

- [Patch Command](#patch-command)
  - [How it works](#how-it-works)
  - [Usage](#usage)
  - [Quick Start](#quick-start)
    - [Pre-requisites](#pre-requisites)
    - [Example Steps](#example-steps)
  - [Limitations](#limitations)

## How it works

- Kubescape already supports scanning the images running in your cluster and stores the image vulnerability result as CRDs in your cluster. More details about the CRDs can be found [here](https://hub.armosec.io/docs/vulnerabilities-relevancy#overview).
- The patch command reads the CRDs for the image vulnerability report and patches the specified image with vulnerabilities.
- The patch command uses [copa](https://github.com/project-copacetic/copacetic) and [buildkit](https://github.com/moby/buildkit) under the hood for patching the container images.

## Usage

```bash
kubescape patch [flags]
```

The patch command can be run in 2 ways:
1. **With sudo privileges**

    You will need to start `buildkitd` if it is not already running
  
    ```bash
    sudo buildkitd & 
    sudo kubescape patch --image <image-name> --kubeconfig <kubeconfig-file-path>
    ```


    > **Note:** When running kubescape with sudo privileges, you will need to specify the kubeconfig file path using the `--kubeconfig` flag. This is because the `buildkitd` daemon runs as root, and does not have access to the kubeconfig file of the user.

2. **Without sudo privileges**

   ```bash
    export BUILDKIT_VERSION=v0.11.4
    export BUILDKIT_PORT=8888

    docker run \
        --detach \
        --rm \
        --privileged \
        -p 127.0.0.1:$BUILDKIT_PORT:$BUILDKIT_PORT/tcp \
        --name buildkitd \
        --entrypoint buildkitd \
        "moby/buildkit:$BUILDKIT_VERSION" \
        --addr tcp://0.0.0.0:$BUILDKIT_PORT

    kubescape patch \
        -i <image-name> \
        -a tcp://0.0.0.0:$BUILDKIT_PORT
   ```

## Quick Start

We will demonstrate how to use the patch command with an example of [nginx](https://www.nginx.com/) image.

### Pre-requisites

- Kubescape should be installed **in your cluster**. If not, follow the instructions [here](https://hub.armosec.io/docs/installation-of-armo-in-cluster#scanning-a-cluster)
- [buildkit](https://github.com/moby/buildkit) daemon installed & pathed.
- [docker](https://docs.docker.com/desktop/install/linux-install/#generic-installation-steps) daemon running and CLI installed & pathed.

### Example Steps

1. Download the nginx image
    ```bash
    docker pull nginx:1.22
    ```

2. Run the nginx image in your kubernetes cluster
    ```bash
    kubectl create deployment nginx --image=docker.io/library/nginx:1.22
    ```

3. `[Optional]` View the image vulnerability report
    ```bash
    kubectl get -n kubescape VulnerabilityManifests
    ```

    The output should be something like below. These are the vulnerabilities of various images running in your cluster.
    ```bash
    NAME                                                     CREATED AT
    docker.io-library-nginx-1.22-08f6b7                      2023-08-02T20:26:31Z
    quay.io-kubescape-kubescape-v2.3.7-ea3c26                2023-08-02T15:36:16Z
    quay.io-kubescape-kubevuln-v0.2.99-7c3f8a                2023-08-02T15:36:13Z
    quay.io-kubescape-storage-v0.0.8-0ba4eb                  2023-08-02T15:36:00Z
    ```

    Get the vulnerability report of your specific image by looking at the name. In this case:
    ```bash
    kubectl get -n kubescape VulnerabilityManifests docker.io-library-nginx-1.22-08f6b7 -o json > nginx-vuln.json
    ```

4. Patch the image. 
    
    You will need to start `buildkitd` if it is not already running:

    ```bash
    sudo buildkitd & 
    sudo kubescape patch --image docker.io/library/nginx:1.22 --kubeconfig ~/.kube/config --tag 1.22-patched
    ```

    Alternatively, you can run `buildkitd` in a container, which allows you to run the patch command without sudo privileges:

    ```bash

    export BUILDKIT_VERSION=v0.11.4
    export BUILDKIT_PORT=8888

    docker run \
        --detach \
        --rm \
        --privileged \
        -p 127.0.0.1:$BUILDKIT_PORT:$BUILDKIT_PORT/tcp \
        --name buildkitd \
        --entrypoint buildkitd \
        "moby/buildkit:$BUILDKIT_VERSION" \
        --addr tcp://0.0.0.0:$BUILDKIT_PORT

    kubescape patch \
        -i docker.io/nginx:1.22 \
        -t 1.22-patched \
        -a tcp://0.0.0.0:$BUILDKIT_PORT

    ```
    
    This will export the patched image to your local docker registry with the tag `1.22-patched`.

5. `[Optional]` Verify that the image is patched. Run that image in your cluster and check the vulnerability report again. You should see that the vulnerabilities are fixed.

    Tag the image to add an organization name to it, and then push to docker hub:
    ```bash
    docker tag nginx:1.22-patched anubhavgupta06/nginx:1.22-patched
    ```
    ```bash
    docker push anubhavgupta06/nginx:1.22-patched
    ```
    
    Create deployment with the patched image:

    ```bash
    kubectl create deployment nginx-patched --image=docker.io/anubhavgupta06/nginx:1.22-patched
    ```

    View the image vulnerability reports:
    ```bash
    kubectl get -n kubescape VulnerabilityManifests
    ```

    The output should be something like below. These are the vulnerabilities of various images running in your cluster.
    ```bash
    NAME                                                     CREATED AT
    docker.io-anubhavgupta06-nginx-1.22-patched-4c49e4       2023-08-02T20:32:40Z
    docker.io-library-nginx-1.22-08f6b7                      2023-08-02T20:26:31Z
    quay.io-kubescape-kubescape-v2.3.7-ea3c26                2023-08-02T15:36:16Z
    quay.io-kubescape-kubevuln-v0.2.99-7c3f8a                2023-08-02T15:36:13Z
    quay.io-kubescape-storage-v0.0.8-0ba4eb                  2023-08-02T15:36:00Z
    ```

    Get the vulnerability report of your specific image by looking at the name. In this case:
    ```bash
    kubectl get -n kubescape VulnerabilityManifests docker.io-anubhavgupta06-nginx-1.22-patched-4c49e4 -o json > nginx-vuln-patched.json
    ```

    Compare the vulnerability report of the patched image with the original image. You will see that the vulnerabilities are fixed.


## Limitations

- It patches only the images which are running in the cluster. It cannot explicitly scan and patch images which are not running in the cluster. This is because kubescape only scans the images which are running in the cluster, and does not support explicit scanning of images.
- It can only fix OS-level vulnerability. It cannot fix application-level vulnerabilities. This is a limitation of copa. The reason behind this is that application level vulnerabilities are best suited to be fixed by the developers of the application.