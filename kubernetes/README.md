Install Calico and Contour

Afterwards install the updated contour service

## Kubeadm Init

```
sudo kubeadm init --pod-network-cidr=192.168.178.0/24
```

## Calico
```
From website
```

## Countour
```
kubectl apply -f https://projectcontour.io/quickstart/contour.yaml
```

Afterwards the `02-service-envoy.yaml` needs to be applied to configure the external ips if no load balancer is used.

## KeyCloak

Keycload isused to create the users

Kubeadm upgrade kills the api server manifest

## Cert Manager

```
kubectl apply -f https://github.com/jetstack/cert-manager/releases/download/v1.1.0/cert-manager.yaml
```
 and the apply the selfsigned-ca.yml to make selfsigned clusters

## OIDC

For OIDC for install keycloak.yml and afterwards all Rolebindings


### Test OIDC

```
kubectl oidc-login setup \--oidc-issuer-url=https://keycloak.kubelab.local/realms/kubelab \--oidc-client-id=kubelab \--oidc-client-secret=SECRET --insecure-skip-tls-verify
```

## nfs-provisioner

The files here are modified to fit the puprose of the project. The original files can be found https://github.com/kubernetes-sigs/nfs-subdir-external-provisioner