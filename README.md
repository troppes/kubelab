<img alt="gitleaks badge" src="https://img.shields.io/badge/protected%20by-gitleaks-blue">

# Kubelab

This repositoy contains all files needed to deploy Kubelab. Kubelab is a protoype created for my masters thesis to test out and demonstrate the capabilities of containers for educational use. 

## What exactly does Kubelab?

Kubelab aims to create a computer labratory inside a Kubernetes cluster. For this it has a custom operator, that deploys two different CRDs. One for people and one for classrooms. Furthermoore it includes a web application, which makes it possible to access the created classes and connect per SSH to the allocated Pods. It is also possible to log in as a teacher to upload files to a class share and control the Pods of their students.

## How to deploy Kubelab?

1. Deploy the cluster with the help of the manifests located in the `manifest` folder.
2. Deploy the Operator and Web-Application, which can be done using the Playbooks found in the `playbooks` Folder.
3. Deploy the Custom Resources to create students, teachers and classes. Either use the example found in the `manifest` folder or for easier usage, the are management playbooks included.
