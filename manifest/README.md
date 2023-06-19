# Setup

## Kubeadm Init

```
sudo kubeadm init --pod-network-cidr=192.168.178.0/24
```

## Calico
```
From website and then apply the custom-ressources
```

For Version 1.30.0 exists a bug, that destroys the garbage collection, to fix you need to set the `bgpfilters` permission in the clusterrole `calico-crds`

k edit clusterrole calico-crds


## Countour
```
kubectl apply -f https://projectcontour.io/quickstart/contour.yaml
```

Afterwards the `02-service-envoy.yaml` needs to be applied to configure the external ips if no load balancer is used.


## Cert Manager

```
kubectl apply -f https://github.com/jetstack/cert-manager/releases/download/v1.1.0/cert-manager.yaml
```
and the apply the selfsigned-ca.yml to make selfsigned clusters

## KeyCloak

Configuration of Keycloak can be found in the thesis.
If the manifest ist upgraded during the kubeadm upgrade process, the api-server needs to be redeployed.

### Test OIDC-Connection

```
kubectl oidc-login setup \--oidc-issuer-url=https://keycloak.kubelab.local/realms/kubelab \--oidc-client-id=kubelab \--oidc-client-secret=SECRET --insecure-skip-tls-verify
```

## nfs-provisioner

The files here are modified to fit the puprose of the project. The original files can be found https://github.com/kubernetes-sigs/nfs-subdir-external-provisioner
They just need to be applied to the cluster and can be tested with the files in the example folder.