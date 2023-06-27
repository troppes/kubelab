# Container

This is a example container for the Kubelab system. It uses a custom entrypoint to create the user folders in the structure needed to work. To create your own image, please refer to the [Ansible-Playbooks](../playbooks/README.md) section. There is a playbook to insert the script correctly into your own image. Please note that the image given needs to be unminimalized, as it needs to have the `bash` command available.

## X11 Forwarding

to enable X11 forwarding dont forget to set the `DISPLAY` variable in your Operating system. In WSL2 the variable is automatically set!