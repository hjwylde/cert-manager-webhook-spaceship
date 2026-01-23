<p align="center">
  <img src="https://raw.githubusercontent.com/hjwylde/cert-manager-webhook-spaceship/main/assets/cert-manager-logo.png" height="256" width="256" alt="cert-manager logo" />
  <img src="https://raw.githubusercontent.com/hjwylde/cert-manager-webhook-spaceship/main/assets/spaceship-logo.png" height="256" width="256" alt="spaceship logo" />
</p>

# ACME Webhook - Spaceship

This solver is used as part of the [cert-manager ACME issuer](https://cert-manager.io/docs/configuration/acme/); it is 
an out-of-tree webhook implementation that can handle DNS01 challenges for [Spaceship](https://www.spaceship.com/) 
domains.

## Requirements

* [Kubernetes](https://kubernetes.io/)
* [cert-manager](https://cert-manager.io/)

It is expected that you are familiar with Kubernetes and cert-manager already, and have a cluster running with 
cert-manager pre-installed.

## Installation

The installation steps assume that cert-manager is running in the `cert-manager` namespace. If this is not the case, 
then adjust the steps to ensure the webhook is installed into the same namespace as cert-manager.

### Manifest

```bash
kubectl apply -f https://raw.githubusercontent.com/hjwylde/cert-manager-webhook-spaceship/v0.1.2/config/manifest.yaml
```

### Kustomize

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
  - github.com/hjwylde/cert-manager-webhook-spaceship/config?ref=v0.1.2
```

### Helm

```bash
helm repo add cert-manager-webhook-spaceship https://hjwylde.github.io/cert-manager-webhook-spaceship
helm install --namespace cert-manager cert-manager-webhook-spaceship cert-manager-webhook-spaceship/cert-manager-webhook-spaceship
```

## Configuration

### Issuer

Create a `ClusterIssuer` or `Issuer` resource. N.B.,

* This example uses the Let's Encrypt staging URL; you will need to replace it with the production URL once the webhook 
  is set up and working
* You must replace the example email address with your own email address
* The `groupName` and `solverName` do not need to be modified

```yaml
---
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-staging
spec:
  acme:
    server: https://acme-staging-v02.api.letsencrypt.org/directory
    email: user@example.com
    profile: tlsserver
    privateKeySecretRef:
      name: letsencrypt-staging
    solvers:
      - dns01:
          webhook:
            groupName: cert-manager-webhook-spaceship.hjwylde.github.io
            solverName: spaceship
            config:
              apiKeySecretRef:
                name: spaceship-credentials
                namespace: cert-manager
                key: api-key
              apiSecretSecretRef:
                name: spaceship-credentials
                namespace: cert-manager
                key: api-secret
```

### Credentials

Create a secret that contains your Spaceship credentials. To get the required credentials:
* Visit [spaceship.com](https://spaceship.com), and log in
* Navigate to the API Manager application
* Create a new API key with permissions for "DNS Records - Write"; you'll need both the API key and secret
* Encode the API key and secret in base64 (`printf "%s" "<api-key>" | base64`; `printf "%s" "<api-secret>" | base64`)

```yaml
---
apiVersion: v1
kind: Secret
metadata:
  name: spaceship-credentials
  namespace: cert-manager
data:
  api-key: <base64-api-key>
  api-secret: <base64-api-secret>
```

### Certificate

Finally, you can create a certificate using your [preferred approach](https://cert-manager.io/docs/usage/certificate/).
