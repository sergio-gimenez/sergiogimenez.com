---
title: "Deploying a KinD Cluster with External Access"
date: 2025-07-03
lastmod: 2025-07-06
categories:
    - "articles"
tags:
    - "kind"
    - "kubernetes"
---

Recently, I needed to deploy a KinD (Kubernetes in Docker) cluster that would be accessible remotely. I had done this before, but never documented the process—so here’s a simple, step-by-step guide for future reference.

{{< alert "circle-info" >}}
 I’m running [clabernetes](https://containerlab.dev/manual/clabernetes/quickstart/), but that’s completely irrelevant for this guide.
{{< /alert >}}

## 1. Create a KinD Cluster Config


First install docker, follow the [official docs](https://docs.docker.com/engine/install/ubuntu/)

Next, install kind following the [official docs](https://kind.sigs.k8s.io/docs/user/quick-start/)

For linux, it's just

```bash
[ $(uname -m) = x86_64 ] && curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.29.0/kind-linux-amd64
chmod +x ./kind
sudo mv ./kind /usr/local/bin/kind
```
Then, create a configuration file for your KinD cluster. This file will define the cluster's networking settings and node roles.

Save the following content as `kind-config.yaml`:

```yaml
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
networking:
  apiServerAddress: "0.0.0.0"
nodes:
  - role: control-plane
    kubeadmConfigPatches:
      - |
        kind: ClusterConfiguration
        apiServer:
          certSANs:
          - "172.27.13.254" # <-- REPLACE with your host IP
          - "localhost"
          - "127.0.0.1"
    extraPortMappings:
      - containerPort: 6443
        hostPort: 6443
        listenAddress: "0.0.0.0"
        protocol: TCP
  - role: worker
  - role: worker
containerdConfigPatches:
  - |-
    [plugins."io.containerd.grpc.v1.cri".containerd]
      discard_unpacked_layers = false
```

## 2. Create the Cluster

Run:

```bash
kind create cluster --name c9s --config kind-config.yaml
```

Once the cluster is up, run `docker ps` on your host. You should see the `c9s-control-plane` container, and its `PORTS` section should include `0.0.0.0:6443->6443/tcp`. You might also see a `127.0.0.1:RANDOM_PORT->6443/tcp`—this is KinD’s default internal mapping, which you can ignore.

## 3. Export the Kubeconfig

```bash
kind get kubeconfig --name c9s > c9s-external-kubeconfig.yaml
```

Now, **edit `c9s-external-kubeconfig.yaml`**. Find the `server:` line (it will likely point to `https://0.0.0.0:6443`). Change it to use **your host IP** and port **6443**.

## 4. Test Your Setup

To quickly test:

```bash
cp c9s-external-kubeconfig.yaml ~/.kube/config
kubectl get nodes
```

You should see output similar to:

```
NAME                STATUS   ROLES           AGE     VERSION
c9s-control-plane   Ready    control-plane   6m3s    v1.33.1
c9s-worker          Ready    <none>          5m48s   v1.33.1
c9s-worker2         Ready    <none>          5m48s   v1.33.1
```

That’s it! You now have a KinD cluster accessible from outside your host.

