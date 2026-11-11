# Single-command Helm prerequisites

These assets are included by build-one-command-package.py in the delivery chart.
The source chart keeps prerequisites.enabled=false for existing administrator-managed installations.

- Operator manifests and 16 CRDs: rendered from charts/ob-operator, version 2.3.4,
  namespace oceanbase-system, matched published test images.
- cert-manager manifests and 6 CRDs: upstream cert-manager v1.12.4 release manifest,
  previously used in the test environments. Images are mirrored without modification.
- prepare.sh: pre-install/pre-upgrade hook; uses a short-lived service-account
  binding to install or reuse shared cluster components. It removes its own binding
  on exit. A Job that cannot start requires retry or explicit binding removal.
- Shared infrastructure is not owned by the application release and is retained.
  Administrators maintain and upgrade it explicitly. CRDs follow Helm's install-only
  CRD behavior; this package never upgrades or deletes existing CRDs automatically.

Do not run concurrent first installations. Offline archives must include the
bootstrap image and embedded manifest images, in addition to application images.
