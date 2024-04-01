# trouble shooting

This document describes how to solve some common errors

## openhydra installation interrupted

* This situation is mostly unrecoverable automatically. We need the following steps to completely reset and reinstall

```bash
# Login to the server
# Switch to the root account
# Start reset
$ systemctl stop kubelet
$ systemctl disable kubelet
# stop and disable openhydra installation service
# because the systemd service will run automatically after the server reboot if  /etc/kubernetes is not exists
# so to avoid chaos, we need to stop and disable the service to make thing easier
$ systemctl stop maas
$ systemctl disable maas
# If the kubeadm reset -f command is stuck for more than 2 minutes, just restart the server
$ kubeadm reset -f
# Clean up the kubeclipper database
$ kcctl clean -Af
# Check for zombie containers, if you reboot you server previously then most likely you don't have any zombie containers remaining just skip this step
$ ctr -n k8s.io container list
# If the above command returns some containers, you need to restart the server or you can skip rebooting server
# because clear all the containers and task for containerd is really a big job, so to easy things up, just reboot the server
$ reboot
# reset again
$ kubeadm reset -f
# remove kubernetes configuration
$ rm -rf /etc/kubernetes

# Start reinstallation of openhydra by rest
$ systemctl start maas

# check the log
$ journalctl -u maas -f
```

## How to update the image of open-hydra server

* Login to the server running open-hydra server and run the following command

```bash
# Manually download the image
$ ctr -n k8s.io i pull docker.io/99cloud/open-hydra-server:latest
# Restart open-hydra server
$ kubectl scale deployment open-hydra-server --replicas=0 -n open-hydra
# Wait for 3 seconds
$ kubectl scale deployment open-hydra-server --replicas=1 -n open-hydra
```
