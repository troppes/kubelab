# Ansible-Playbooks

## Install Kubelab

To install Kubelab just run the `00_install_kubelab.yml` playbook. It automatically creates two Namespaces and the deployments to run Kubelab.

## List CRDS

```
10_list_users.yml
20_list_classes.yml
```

Run those files to see the current state of the Kubelab custom resources.

## Manage CRDS

```
11_manage_users.yml
21_manage_classes.yml
```

Those two files allow for creating, updating and deleting the CRDS. For this entries can be made in the classes.csv and users.csv files, in the root folder. To create or update just run the playbook, to delete add `-e=delete=1` to the command.


## Create own Docker Image

To create your own compatible version of the Kubelab Docker Image, just run the `30_build_docker_image.yml` playbook. It will build the image and push it to the Docker Hub, with the needed configurations made. You can change the image name and tag, as well as all other settings in the `group_vars/all.yml_dist`, which needs to be renamed to `group_vars/all.yml`.
Please note, that in the image given, the entrypoint or cmd will be overwritten to use the script, that is needed for the setup. If you want to run any scripts, please use a simple run command in your base image. Please note that the image given needs to be unminimalized, as it needs to have the `bash` command available.