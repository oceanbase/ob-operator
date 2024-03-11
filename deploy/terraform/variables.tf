variable "k8s_host" {
  type        = string
  description = "The hostname (in form of URI) of the Kubernetes API."
}

variable "k8s_client_certificate" {
  type        = string
  description = "PEM-encoded client certificate for TLS authentication."
}

variable "k8s_client_key" {
  type        = string
  description = "PEM-encoded client certificate key for TLS authentication."
}

variable "k8s_cluster_ca_certificate" {
  type        = string
  description = "PEM-encoded root certificates bundle for TLS authentication."
}
