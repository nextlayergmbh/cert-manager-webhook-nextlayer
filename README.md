# ACME Webhook for next layer DNS

This project provides a [cert-manager](https://cert-manager.io) ACME Webhook for [next layer DNS](https://www.nextlayer.at/) 
and is based on the [Example Webhook](https://github.com/jetstack/cert-manager-webhook-example).

## Requirements
* helm >= v3.0.0
* kubernetes >= v1.14.0
* cert-manager >= 0.12.0

## Configuration

The following table lists the configurable parameters of the cert-manager chart and their default values.

| Parameter | Description | Default |
| --------- | ----------- | ------- |
| `groupName` | Group name of the API service. | `dns.nextlayer.at` |
| `certManager.namespace` | Namespace where cert-manager is deployed to. | `kube-system` |
| `certManager.serviceAccountName` | Service account of cert-manager installation. | `cert-manager` |
| `image.repository` | Image repository | `registry.nextlayer.at/nextlayer/cert-manager-webhook-nextlayer` |
| `image.tag` | Image tag | `latest` |
| `image.pullPolicy` | Image pull policy | `Always` |
| `service.type` | API service type | `ClusterIP` |
| `service.port` | API service port | `443` |
| `resources` | CPU/memory resource requests/limits | `{}` |
| `nodeSelector` | Node labels for pod assignment | `{}` |
| `affinity` | Node affinity for pod assignment | `{}` |
| `tolerations` | Node tolerations for pod assignment | `[]` |

## Installation

### cert-manager

Follow the [instructions](https://cert-manager.io/docs/installation/) using the cert-manager documentation to install it within your cluster.

### Webhook

#### By cloning the repo
```bash
git clone https://github.com/nextlayergmbh/cert-manager-webhook-nextlayer.git
cd cert-manager-webhook-nextlayer
helm install --namespace cert-manager cert-manager-webhook-nextlayer ./deploy/cert-manager-webhook-nextlayer
```

#### By adding the helm repo
```bash
helm repo add nextlayercm https://nextlayergmbh.github.io/cert-manager-webhook-nextlayer/
helm repo update
helm install --namespace cert-manager nextlayercm/cert-manager-webhook-nextlayer
```


**Note**: The kubernetes resources used to install the Webhook should be deployed within the same namespace as the cert-manager.

To uninstall the webhook run
```bash
helm uninstall --namespace cert-manager cert-manager-webhook-nextlayer
```

## Issuer

Create a `ClusterIssuer` or `Issuer` resource as following:
```yaml
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-staging
spec:
  acme:
    # The ACME server URL
    server: https://acme-staging-v02.api.letsencrypt.org/directory

    # Email address used for ACME registration
    email: mail@example.com # REPLACE THIS WITH YOUR EMAIL

    # Name of a secret used to store the ACME account private key
    privateKeySecretRef:
      name: letsencrypt-staging

    solvers:
      - dns01:
          webhook:
            groupName: dns.nextlayer.at
            solverName: nextlayer
            config:
              APIKey: <YOUR-DNS-API-KEY-HERE>
```

### Credentials

For accessing the next layer DNS API, you need an API Token which you can request via the next layer support. 
Currently we don't provide a way to use secrets for you API KEY.

## Thanks

Thanks to [mecodia GmbH](https://github.com/mecodia/cert-manager-webhook-hetzner) and [Stephan Müller](https://gitlab.com/smueller18/cert-manager-webhook-inwx) whose project served as an example for `cert-manager-webhook-nextlayer`.