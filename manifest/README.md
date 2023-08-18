# Examples & Setup

## Examples

In the `examples` folder, examples for all custom ressources of the protoype are found.

## Kubeadm init

Install kubeadm and set up the cluster. If you want to use the same Networking, that was used for the prototype, use the following command:
```
sudo kubeadm init --pod-network-cidr=192.168.178.0/24
```

## Calico

The network addon chosen for the protoype was Calico, but any Networking addon should work. For the installation please refer to the Calico documentation.

In Version 1.30.0 exists a bug, that destroys the Kubernetes garbage collection. To fix this, you need to set the `bgpfilters` permission in the clusterrole `calico-crds`. This can be done with the help of the `k edit clusterrole calico-crds` command.

To configure it with the Network defined above, please deploy the `custom-ressources.yml` in the `01-calico` folder.

## Cert Manager

Cert-Manger was used to facilitate the certificate generation. For the installtion please refer to the Cert-Manager documentation. The setup includes a resource to install a selfsigned-issuer, making it possible to generate private certificates.

## Countour

Contour was used as the Ingress Controller, but any Ingress Controller sould work. For the installation please refer to the Contour documentation.
After installing use the `02-service-envoy.yaml` to configure Contour to use external IPs. This step is only required for bare bone clusters without and external load balancer.

## KeyCloak

Keycloak is used as the OIDC provider for the webapp. To install Keycloak, either refer to the documentation or use the manifest provided.
It is important to state, that the integration of Keycloak into the cluster needs to be done manually. Helpful pointers can be found in the thesis and an example is within the Keycloak folder. If the manifest ist upgraded during the kubeadm upgrade process, the api-server needs to be redeployed.

### Test OIDC-Connection

To test the OIDC connection to the cluster the following command can be used:
```
kubectl oidc-login setup \--oidc-issuer-url=URL \--oidc-client-id=CLIENT_ID \--oidc-client-secret=SECRET --insecure-skip-tls-verify
```

## NFS-Provisioner

The files here are modified to fit the puprose of the project. The original files can be found https://github.com/kubernetes-sigs/nfs-subdir-external-provisioner. They just need to be applied to the cluster and can be tested with the files in the example folder. Please note, that `03-deployment.yml` needs to be edited with the IP of the NFS server.

## Kubelab Operator

The Kubelab Operator manages the custom resources needed for this protoype. Pleas use `k create` for the installation, since `k apply` leads to an incomplete creation. It is furthermoore recommended to use the playbook provided to install the Operator and the web application.

## Kublab Web

The Kubelab Web server is the main interactive component of the stack, allowing the users to manage their containers. To make it easier to deploy, please refer to the playbook. If not, pleas edit `02-configmap.yml` and `03-kubelab-web.yml` to your liking.
