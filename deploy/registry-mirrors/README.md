# Optional image acceleration without changing image names

These examples let users opt into a registry mirror while retaining the image
names, release tags and digests in the delivery. They are included in both the
Compose and native Helm bundles. Installation does not apply them automatically.

| Deployment runtime | Configuration | Scope of this example |
| --- | --- | --- |
| Docker Engine / Compose | `daemon.json` | Docker Hub only |
| K3s embedded containerd | `registries.yaml` on each pulling node | Docker Hub and Quay; optional GHCR |
| Other Kubernetes runtimes | Runtime-specific configuration | Use your cluster administrator's mirror setup |

For Kubernetes, check the runtime before choosing a configuration:

```bash
kubectl get nodes -o custom-columns=NAME:.metadata.name,RUNTIME:.status.nodeInfo.containerRuntimeVersion
```

`docker://...` uses the Docker instructions below, including K3s installed with
`--docker`. The K3s instructions apply to its embedded `containerd://...` runtime;
other containerd distributions may use different paths and configuration formats.
Helm itself does not choose or configure the runtime.

The Docker Hub configuration retains all 14 requested mirror addresses in their
original order, in both the Docker example and the K3s `docker.io` mapping. The
merge helper adds the complete list by default. No address is filtered out based
on a reachability check. Existing mirrors keep their order when merging.

These are optional third-party endpoints; including an address does not verify
availability of every release image. The separate Quay and optional GHCR mappings
use [DaoCloud's registry-specific endpoints](https://github.com/DaoCloud/public-image-mirror).

A mirror does not publish private or missing release images or grant upstream
repository access. A public release must first publish every image in its
`image-lock.json`, or supply the matching offline image archive. Retain the
release's exact references; do not substitute an unrelated version because it
is cached. The package build does not verify public registry publication.

Commands below run from the extracted delivery directory. In a source checkout,
replace `registry-mirrors/` with `deploy/registry-mirrors/`.

## Docker Engine / Compose

`docker-daemon.json` contains the requested settings: all 14 Docker Hub mirrors,
ten concurrent layer downloads and five concurrent uploads. Docker's
`registry-mirrors` does not cover `quay.io`, `ghcr.io` or arbitrary private
registries. See [Docker's mirror documentation](https://docs.docker.com/docker-hub/image-library/mirror/).

The Python 3 standard-library helper merges settings without replacing other
daemon options. Existing mirrors retain their order, and existing concurrency
settings take precedence. It only adds the sample values where absent.

Preview the merged configuration; this does not write files:

```bash
sudo python3 registry-mirrors/configure-docker.py
```

Apply the previewed settings explicitly:

```bash
sudo python3 registry-mirrors/configure-docker.py --apply
```

The helper backs up an existing file alongside it, preserves its mode and owner,
then replaces it atomically. Repeating the command does not add duplicates. An
invalid JSON file or a symbolic-link config is rejected. It never reloads Docker
or changes containers. If `/etc/docker` does not yet exist, create it first with
`sudo install -d -m 755 /etc/docker`. To add your organization's Docker Hub mirror
instead, pass `--mirror https://YOUR_MIRROR`; repeat the option for multiple
known mirrors. Use `--config PATH` when the daemon uses another config file.

On a Linux host with the standard systemd Docker service, validate and reload:

```bash
sudo dockerd --validate --config-file=/etc/docker/daemon.json && \
  sudo systemctl kill --kill-whom=main --signal=HUP docker.service
docker info --format '{{json .RegistryConfig.Mirrors}}'
```

These options support a SIGHUP reload; restarting containers is unnecessary.
Options supplied in daemon startup flags can conflict with file settings. Check
the service log if the active mirrors did not change, and reconcile those flags
before reloading. See [Docker daemon configuration](https://docs.docker.com/reference/cli/dockerd/#configuration-reload-behavior).

From the Compose delivery directory, verify its actual release images:

```bash
docker compose config --images
docker compose pull
docker compose up -d
```

Quay and GHCR images still need direct access, your existing network proxy, or a
matching offline image archive when using Docker. To undo the file change, restore
the backup path printed by the helper and repeat validation and reload.

## K3s embedded containerd

Helm installs resources; image downloads are performed by each node's runtime.
Docker's `daemon.json` does not configure K3s embedded containerd.

On **each node that pulls images**, including schedulable server nodes, merge
`k3s-registries.yaml` into `/etc/rancher/k3s/registries.yaml`. Keep existing
`mirrors`, `configs`, authentication, TLS settings and rewrites. Back up an
existing file before editing it; do not replace it with this example.

For a node with **no existing file**, install the example:

```bash
sudo install -d -m 755 /etc/rancher/k3s
sudo test ! -e /etc/rancher/k3s/registries.yaml && \
  sudo install -m 600 registry-mirrors/k3s-registries.yaml /etc/rancher/k3s/registries.yaml
```

The `docker.io` and `quay.io` mappings preserve original image names. The
commented GHCR mapping is available for releases whose public GHCR images have
also been checked against the mirror's allowlist. Enable it only when needed.

Apply the configuration on each node by restarting its K3s service. Use the
appropriate command for that node's role, and handle existing clusters one node
at a time according to their maintenance procedure:

```bash
sudo systemctl restart k3s        # server node
# OR
sudo systemctl restart k3s-agent  # agent node
```

K3s normally falls back to the original registry after mirror endpoints. This
example preserves that behavior. See [K3s registry configuration](https://docs.k3s.io/installation/private-registry).

Test every exact image reference from the delivery's `image-lock.json` on the
target workers with `sudo k3s crictl pull IMAGE_REFERENCE`, then follow the Helm
bundle's `QUICKSTART.md`. A cached image or direct-registry fallback can make a
pull succeed without using the mirror; inspect
`/var/lib/rancher/k3s/agent/containerd/containerd.log` if you need to confirm the
endpoint used. Do not treat a successful HTTP request to the mirror homepage as
release-image validation.
