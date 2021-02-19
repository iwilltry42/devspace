---
title: "Command - devspace use namespace"
sidebar_label: devspace use namespace
---


Tells DevSpace which namespace to use

## Synopsis


```
devspace use namespace [flags]
```

```
#######################################################
############## devspace use namespace #################
#######################################################
Set the default namespace to deploy to

Example:
devspace use namespace my-namespace
#######################################################
```


## Flags

```
  -h, --help    help for namespace
      --reset   Resets the default namespace of the current kube-context
```


## Global & Inherited Flags

```
      --config string            The devspace config file to use
      --debug                    Prints the stack trace if an error occurs
      --kube-context string      The kubernetes context to use
  -n, --namespace string         The kubernetes namespace to use
      --no-warn                  If true does not show any warning when deploying into a different namespace or kube-context than before
  -p, --profile string           The devspace profile to use (if there is any)
      --profile-parent strings   One or more profiles that should be applied before the specified profile (e.g. devspace dev --profile-parent=base1 --profile-parent=base2 --profile=my-profile)
      --profile-refresh          If true will pull and re-download profile parent sources
      --restore-vars             If true will restore the variables from kubernetes before loading the config
      --save-vars                If true will save the variables to kubernetes after loading the config
      --silent                   Run in silent mode and prevents any devspace log output except panics & fatals
  -s, --switch-context           Switches and uses the last kube context and namespace that was used to deploy the DevSpace project
      --var strings              Variables to override during execution (e.g. --var=MYVAR=MYVALUE)
      --vars-secret string       The secret to restore/save the variables from/to, if --restore-vars or --save-vars is enabled (default "devspace-vars")
```
