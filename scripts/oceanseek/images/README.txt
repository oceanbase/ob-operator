OceanSeek component images

Dashboard and helper images are built from the ob-operator sources with the
release-dashboard and release-helper workflows. Their production Dockerfiles
are build/Dockerfile.dashboard and build/Dockerfile.obhelper. Use the version
and platform digest recorded in the selected release. The public obdiag
source is at https://github.com/oceanbase/obdiag.

The optional local helper and observer Dockerfiles take an explicit base image
and local inputs. Build oceanbase-helper and oceanseek-readiness for the target platform with
make helper oceanseek-readiness and place both binaries in the helper build
context. The observer build uses the signed tar RPM identified in Dockerfile.observer.

kubernetes_log_transport.py downloads diagnostic archives in checked chunks.
patch_obdiag.py adapts the matching obdiag 5.0.1 source layout to this transport.
test_kubernetes_log_transport.py covers checksum verification, truncation and
replacement of the output file. The MinIO rejoin build and patch are in
minio-rejoin/.
