terraform {
  required_providers {
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.6.1"
    }

    helm = {
      source  = "hashicorp/helm"
      version = "~> 2.2.0"
    }

    time = {
      source  = "hashicorp/time"
      version = "~> 0.7.2"
    }
  }
}

provider "kubernetes" {
  host = var.k8s_host

  client_certificate     = base64decode(var.k8s_client_certificate)
  client_key             = base64decode(var.k8s_client_key)
  cluster_ca_certificate = base64decode(var.k8s_cluster_ca_certificate)
}

provider "helm" {
  kubernetes {
    host = var.k8s_host

    client_certificate     = base64decode(var.k8s_client_certificate)
    client_key             = base64decode(var.k8s_client_key)
    cluster_ca_certificate = base64decode(var.k8s_cluster_ca_certificate)
  }
}


resource "kubernetes_namespace" "oceanbase-system" {
  lifecycle {
    ignore_changes = [metadata]
  }

  metadata {
    name = "oceanbase-system"
  }
}

resource "kubernetes_namespace" "oceanbase" {
  lifecycle {
    ignore_changes = [metadata]
  }

  metadata {
    name = "oceanbase"
  }
}

resource "time_sleep" "wait_30_seconds" {
  depends_on = [kubernetes_namespace.oceanbase-system]

  destroy_duration = "30s"
}

resource "helm_release" "ob-operator" {
  repository = "https://oceanbase.github.io/ob-operator"
  chart      = "ob-operator"
  name       = "ob-operator"
  namespace  = "oceanbase-system"
  depends_on = [time_sleep.wait_30_seconds]
}
