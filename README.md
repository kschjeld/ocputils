# OpenShift utils

## Introduction

A small batch of helper utils related to OpenShift and the way we use it.

## Installation

To install individual utils, run ```` go install  ```` for the individual tools under cmd/

Example:

````shell script
go install cmd/listocpgroups.go
````

### List ocp groups

List groups and users in OpenShift, including full name if available. Can also list information for a single user
through the _search_ parameter.

### Match image and Git

List all deployment configs in current cluster, and try to extract Docker image information for each running container using metadata label
_git.url_ in the image.

### Extract projectsetups (extractprojectsetups)

This utility will loop through all namespaces/projects in OpenShift and extract Projectsetup for groups of projects that fit together. It 
will filter our openshift-specific namespaces.

Grouping will be done based on name suffix, placing all with common base name and different suffixes into same group. The actual suffixes
used for grouping namepaces must be modified according to standards used (defined in map suffixRoleMappings). 

Output can be written to stdout or as files in a directory.

## Author

kristian@fluxconsulting.no
