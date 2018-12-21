# Namespace Cleaner

This is a simple project that will clean Kubernetes' namespaces that you don't need.

## Why this project?

Kubernetes is a very flexible system and it's often used in multi tenant environments. Kubernetes itself does not prescribe how to use namespaces
and there are not many recommendations out there on how to use them or how to map them to organizations and teams.
There are cases in which we can use multiple namespaces to test things and be able to delete them quickly. Or your organization is just not that disciplinate at using namespaces. What happens is that you end up with plenty of namespaces lying around which cause confusion and consume resources if there are things deployed and forgotten into them. 
This project wants to be a part of code that allows you to get rid of unwanted namespaces.

## How it works.

This code is supposed to be run as a Job or CronJob periodically. The code makes a single pass over the namespaces, exclude some namespaces that we should never delete (`default`, `kube-system` and `kube-public`) and delete all the other namespaces. The user can provide a flag `namespaces-to-retain` to specify other namespaces to **not** delete.
Please note that the default behaviour of this code is to **not** delete namespaces unless the flag `--yes` is specified. This is done to prevent mistaskes.

## Deploy

You can find a sample manifest in [this file](cronjob.yaml).