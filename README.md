# Kubelab

This repo contains all the files to deploy the kubelab project. 

## What is Kubelab?

Kubelab aims to create a computer labratory inside a Kubernetes cluster. For this it has a custom operator, that deploys two different CRDs. One for people and one for classrooms. Further Information can be found in the thesis (link will follow). 

## How to deploy Kubelab?

First you need to use the manifest/setup readme to deploy all setup components. Afterwards you need to deploy the operator and lastly the web server. 