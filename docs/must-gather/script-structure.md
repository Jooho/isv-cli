# Must Gather Structure

This show you how the must gather structure looks like and also output as well.
**must-gather feature is still in heavy develpment so this doc might not be synced up to the latest.**

## Script Structure


- (F)Dockerfile [ex1](https://github.com/openshift/ocs-operator/blob/master/must-gather/Dockerfile), [ex2](https://github.com/openshift/local-storage-operator/Dockerfile.mustgather.rhel7)
- (F)Makefile [ex1](https://github.com/openshift/local-storage-operator/Makefile)
  - build
  - test
    - functest
- (D)collection-scripts
  - (F)gather
    - Container image `Endpoint`. Trigger must-gather
  - (F)version
    - Indicates the product (first line) and the version (second line, major.minor.micro.qualifier)
- (D)functests
  - (F)functests.sh
    - Executes must-gather image and produces tar ball.
  - (F)output-check.sh
    - Check if expected output files exist and they are not empty.
- (D)templates (optional)
  - (F)pod.template
    - If pod yaml needed, this template file can be used.
- (D)output (optional)
  - (F)output-contents
    - List up output files and folders with explanation.



### Scripts 
- [gather](../../examples/must-gather/collection-scripts/gather)
  - Entrypoint script
  - Usually, it should not be touched by anybody.
- [gather_namespace_resource](../../examples/must-gather/collection-scripts/gather_namespaced_resources)
  - Namespace level data
  - Managed by Red Hat
  - Admin user can gather data
  - [Target Objects](../../examples/must-gather/collection-scripts/gather_namespaced_resources)
- [gather_operand_resource](../../examples/must-gather/collection-scripts/gather_operand_resources)
  - Operand specific data
  - Managed by ISV
  - Admin role must be enough to get this data 
- [commons.sh](../../examples/must-gather/configs/commons.sh)
  - Common variables
  - *2 Mandatory update variables*
    - **OPERATOR-NAME**
      - Liternally, it is for ISV operator name
    - **CUSTOM_RESOURCE_LIST**
      - CR list that operator can create.

- [Makefile](../../examples/must-gather/Makefile)
  - 4 Macros
    - check-must-gather
      - Check shell scripts
      - Before you build/push must-gather image, you must check all scripts
    - build-must-gather
      - Build a must-gather image
      - Verify must-gather image version with operator version
    - push-must-gather
      - Push must the must-gather image 
      - Make sure that you should login the repostiory first
    - must-gather
      - Combined macro `build-must-gather` and `push-must-gather`
  - Mandatory update part
    ~~~
    OPERATOR_NAME ?=nfs-provisioner-operator
    DEFAULT_IMAGE_REGISTRY=quay.io
    DEFAULT_REGISTRY_NAMESPACE=jooholee
    DEFAULT_IMAGE_TAG=0.0.1
    ~~~
- [Dockerfile.rhel](../../examples/must-gather/Dockerfile.rhel)
  - Default Dockerfile for a must-gather 
  

# Output Structure
## No tar 
~~~
 must-gather.local.RAND/
 - timestamp
 - event-filter.html
 - quay-io-%IMAGAE%-sha256-XXXX/
   - gather-debug.log
   - namespace-scoped-resources/ 
     - event_filter_data/
     - oc_get/       
     - oc_desc/ 
     - yaml/
     - pod_log/
       - containers/
   - operand/
     - Freeform
~~~
## Tar
event-filter.html is not generated automatically but with the data, you can generate it
~~~
 must-gather.local.RAND/
 - timestamp
 - must-gather.tar
~~~