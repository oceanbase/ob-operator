#!/usr/bin/env python3
"""Prepare the shared Operator using the invoking administrator's kubeconfig."""
import argparse
import hashlib
import json
from pathlib import Path
import re
import subprocess
import time

OWNER = "oceanbase.oceanbase.com/lakehouse-bootstrap-manifest"
NAMESPACE = "oceanbase-system"
DEPLOYMENT = "oceanbase-controller-manager"
SERVICE = "oceanbase-webhook-service"
LAKEHOUSE_CRDS = {
    "obclusters.oceanbase.oceanbase.com": ("deploymentMode", "logServiceRef", "sharedStorageInfo"),
    "oblogserviceclusters.oceanbase.oceanbase.com": ("objectStoreUrl", "logService"),
}
CERT_DEPLOYMENTS = ("cert-manager", "cert-manager-cainjector", "cert-manager-webhook")


class BootstrapError(RuntimeError):
    pass


class Kubectl:
    def __init__(self, executable, context, kubeconfig):
        self.command = [executable]
        if context:
            self.command += ["--context", context]
        if kubeconfig:
            self.command += ["--kubeconfig", kubeconfig]

    def run(self, *args, obj=None, timeout=30):
        result = subprocess.run(
            [*self.command, f"--request-timeout={timeout}s", *args],
            input=json.dumps(obj) if obj is not None else None,
            capture_output=True, text=True, timeout=timeout + 10,
        )
        if result.returncode:
            raise BootstrapError(result.stderr.strip() or "kubectl command failed: " + " ".join(args))
        return result.stdout

    def get(self, *args):
        raw = self.run("get", *args, "--ignore-not-found", "-o", "json")
        return json.loads(raw) if raw.strip() else {}

    def get_object(self, obj):
        raw = self.run("get", "-f", "-", "--ignore-not-found", "-o", "json", obj=obj)
        if not raw.strip():
            return {}
        document = json.loads(raw)
        # kubectl get -f may return a List even for one object.
        if document.get("kind") == "List":
            items = document.get("items", [])
            if len(items) > 1:
                raise BootstrapError(f"Expected at most one result for {object_name(obj)}")
            return items[0] if items else {}
        return document

    def create(self, obj):
        self.run("create", "-f", "-", obj=obj)


def object_name(obj):
    meta = obj["metadata"]
    name = meta["name"]
    if meta.get("namespace"):
        name = meta["namespace"] + "/" + name
    return f"{obj['kind']} {name}"


def owner(obj):
    return obj.get("metadata", {}).get("annotations", {}).get(OWNER)


def load_yaml_objects(paths):
    # Source checkouts and legacy development payloads still use YAML.
    try:
        import yaml
    except ModuleNotFoundError as error:
        raise BootstrapError("Use the packaged Chart for administrator preparation, or install PyYAML for a source checkout") from error
    try:
        return [doc for path in paths for doc in yaml.safe_load_all(path.read_bytes()) if doc]
    except yaml.YAMLError as error:
        raise BootstrapError(f"Invalid Operator YAML: {error}") from error


def load_objects(payload):
    crds = sorted((payload / "crds").glob("*.yaml"))
    # Older source trees still carry historical cert-manager manifests. Only
    # OceanBase's APIs belong to this bootstrap.
    crds = [p for p in crds if ".oceanbase.com." in p.name]
    operator = payload / "operator.yaml"
    if not operator.is_file():
        operator = payload / "manifests/operator.yaml"
    paths = [*crds, operator]
    digest = hashlib.sha256()
    for path in paths:
        source = path.read_bytes()
        digest.update(path.name.encode() + b"\0" + source + b"\0")
    normalized = payload / "objects.json"
    objects = json.loads(normalized.read_text()) if normalized.is_file() else load_yaml_objects(paths)
    if not isinstance(objects, list) or any(not isinstance(obj, dict) for obj in objects):
        raise BootstrapError("Invalid packaged Operator objects")
    version = digest.hexdigest()
    if not set(LAKEHOUSE_CRDS) <= {
        obj["metadata"]["name"] for obj in objects if obj["kind"] == "CustomResourceDefinition"
    }:
        raise BootstrapError("Operator payload does not contain the required Lakehouse CRDs")
    for obj in objects:
        obj["metadata"].setdefault("annotations", {})[OWNER] = version
    return objects, version


def require_owned(current, desired, version):
    if current and owner(current) != version:
        raise BootstrapError(
            f"{object_name(desired)} belongs to another installation or artifact. "
            "Finish or repair it as its administrator, then use --verify-only."
        )


def ensure_object(kube, obj, version):
    current = kube.get_object(obj)
    if current:
        require_owned(current, obj, version)
        return
    try:
        kube.create(obj)
    except BootstrapError:
        # A simultaneous run with the exact artifact may have won create.
        current = kube.get_object(obj)
        if not current:
            raise
        require_owned(current, obj, version)


def remaining(deadline):
    seconds = int(deadline - time.monotonic())
    if seconds <= 0:
        raise BootstrapError("Readiness timed out; inspect shared components and retry")
    return seconds


def wait(kube, deadline, *args):
    timeout = remaining(deadline)
    kube.run(*args, f"--timeout={timeout}s", timeout=timeout)


def wait_for(check, deadline, message):
    while time.monotonic() < deadline:
        if check():
            return
        time.sleep(2)
    raise BootstrapError(message)


def ca_ready(obj):
    hooks = obj.get("webhooks", [])
    return bool(hooks) and all(hook.get("clientConfig", {}).get("caBundle") for hook in hooks)


def wait_cert_manager(kube, namespace, deadline):
    for name in CERT_DEPLOYMENTS:
        if not kube.get("-n", namespace, "deployment", name):
            raise BootstrapError(f"Install a supported cert-manager in {namespace} before preparing the Operator")
        wait(kube, deadline, "-n", namespace, "rollout", "status", f"deployment/{name}")
    wait_for(lambda: ca_ready(kube.get("validatingwebhookconfiguration", "cert-manager-webhook")),
             deadline, "cert-manager webhook CA did not become ready")


def verify_operator(kube, namespace, deadline):
    for name, fields in LAKEHOUSE_CRDS.items():
        wait(kube, deadline, "wait", "--for=condition=Established", f"crd/{name}")
        crd = kube.get("crd", name)
        api = next((v for v in crd.get("spec", {}).get("versions", []) if v["name"] == "v1alpha1"), {})
        if not api.get("served"):
            raise BootstrapError(f"{name} must serve v1alpha1")
        schema = api.get("schema", {}).get("openAPIV3Schema", {})
        properties = schema.get("properties", {}).get("spec", {}).get("properties", {})
        for field in fields:
            if not properties.get(field, {}).get("type"):
                raise BootstrapError(f"{name} is missing spec.{field}")
    wait(kube, deadline, "-n", namespace, "rollout", "status", f"deployment/{DEPLOYMENT}")

    def webhooks_ready():
        slices = kube.get("-n", namespace, "endpointslices", "-l", f"kubernetes.io/service-name={SERVICE}")
        if not any(endpoint.get("addresses") and endpoint.get("conditions", {}).get("ready") is True
                   for item in slices.get("items", []) for endpoint in item.get("endpoints", [])):
            return False
        return all(ca_ready(kube.get(kind, f"oceanbase-{prefix}-webhook-configuration"))
                   for kind, prefix in (("validatingwebhookconfiguration", "validating"),
                                        ("mutatingwebhookconfiguration", "mutating")))

    wait_for(webhooks_ready, deadline, "OceanBase Operator webhook endpoints or CA did not become ready")


def prepare(kube, objects, version, cert_namespace, timeout, verify_only):
    deadline = time.monotonic() + timeout
    deployments = kube.get("deployments", "-A", "--field-selector", f"metadata.name={DEPLOYMENT}").get("items", [])
    if len(deployments) > 1:
        raise BootstrapError("Multiple OceanBase Operator Deployments exist; choose the cluster owner before proceeding")
    existing = deployments[0] if deployments else None
    if verify_only or (existing and owner(existing) != version):
        if not existing:
            raise BootstrapError("No administrator-managed OceanBase Operator Deployment was found")
        namespace = existing["metadata"]["namespace"]
        verify_operator(kube, namespace, deadline)
        print(f"Existing OceanBase Operator and Lakehouse APIs are ready in {namespace}.")
        return

    # Wait for discovery of cert-manager kinds before reading Operator objects.
    wait_cert_manager(kube, cert_namespace, deadline)
    # Detect partial objects from other installers before creating anything.
    for obj in objects:
        require_owned(kube.get_object(obj), obj, version)
    if not kube.get("namespace", NAMESPACE):
        try:
            kube.create({"apiVersion": "v1", "kind": "Namespace", "metadata": {"name": NAMESPACE}})
        except BootstrapError:
            if not kube.get("namespace", NAMESPACE):
                raise
    for obj in objects:
        ensure_object(kube, obj, version)
        if obj["kind"] == "CustomResourceDefinition":
            wait(kube, deadline, "wait", "--for=condition=Established", f"crd/{obj['metadata']['name']}")
    wait(kube, deadline, "-n", NAMESPACE, "wait", "--for=condition=Ready", "certificate/oceanbase-serving-cert")
    verify_operator(kube, NAMESPACE, deadline)
    print(f"OceanBase Operator and Lakehouse APIs are ready in {NAMESPACE}.")


def revoke_legacy_binding(kube, namespace, release):
    if not re.fullmatch(r"[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?", namespace):
        raise BootstrapError("Expected a Kubernetes namespace name")
    if len(release) > 53 or not re.fullmatch(r"[a-z0-9](?:[a-z0-9.-]*[a-z0-9])?", release):
        raise BootstrapError("Expected a Helm release name")
    service_account = release + "-prerequisites"
    name = namespace + "-" + service_account
    binding = kube.get("clusterrolebinding", name)
    if not binding:
        print(f"Legacy binding {name} is already absent.")
        return
    metadata = binding["metadata"]
    role = {"apiGroup": "rbac.authorization.k8s.io", "kind": "ClusterRole", "name": "cluster-admin"}
    expected = {"apiGroup": "", "kind": "ServiceAccount", "name": service_account, "namespace": namespace}
    subjects = [{"apiGroup": "", **subject} for subject in binding.get("subjects", [])]
    if (metadata["name"] != name or binding.get("roleRef") != role or subjects != [expected]
            or metadata.get("annotations", {}).get("helm.sh/hook") != "pre-install,pre-upgrade"):
        raise BootstrapError(f"Refusing to delete {name}: it does not match the legacy preparation hook")
    job = kube.get("-n", namespace, "job", service_account)
    if job and not any(c.get("status") == "True" and c.get("type") in {"Complete", "Failed"}
                       for c in job.get("status", {}).get("conditions", [])):
        raise BootstrapError("The legacy preparation Job is active. Stop or finish it before revoking access.")
    pods = kube.get("-n", namespace, "pods", "-l", f"job-name={service_account}")
    if any(p.get("status", {}).get("phase") not in {"Succeeded", "Failed"} for p in pods.get("items", [])):
        raise BootstrapError("A legacy preparation Pod is active. Stop or finish it before revoking access.")
    kube.run("delete", "--raw", f"/apis/rbac.authorization.k8s.io/v1/clusterrolebindings/{name}", "-f", "-",
             obj={"apiVersion": "v1", "kind": "DeleteOptions",
                  "preconditions": {"uid": metadata["uid"], "resourceVersion": metadata["resourceVersion"]}})
    print(f"Revoked verified legacy preparation binding {name}.")


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--payload", type=Path, default=Path(__file__).resolve().parent)
    parser.add_argument("--kubectl", default="kubectl")
    parser.add_argument("--kubeconfig")
    parser.add_argument("--context")
    parser.add_argument("--cert-manager-namespace", default="cert-manager")
    parser.add_argument("--timeout", type=int, default=1800, help="Shared readiness budget in seconds")
    action = parser.add_mutually_exclusive_group()
    action.add_argument("--verify-only", action="store_true", help="Check existing Operator APIs and webhook")
    action.add_argument("--revoke-legacy-binding", nargs=2, metavar=("NAMESPACE", "RELEASE"),
                        help="Explicitly revoke a verified inactive legacy preparation binding")
    args = parser.parse_args()
    if args.timeout < 60 or args.timeout > 3600:
        parser.error("--timeout must be between 60 and 3600 seconds")
    try:
        kube = Kubectl(args.kubectl, args.context, args.kubeconfig)
        if args.revoke_legacy_binding:
            revoke_legacy_binding(kube, *args.revoke_legacy_binding)
        else:
            objects, version = load_objects(args.payload)
            prepare(kube, objects, version, args.cert_manager_namespace, args.timeout, args.verify_only)
    except (BootstrapError, OSError, ValueError, KeyError, subprocess.TimeoutExpired) as error:
        parser.exit(1, f"Operator bootstrap failed: {error}\n")


if __name__ == "__main__":
    main()
