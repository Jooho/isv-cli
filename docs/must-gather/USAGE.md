
# ISV CLI Usage
Let's check it out what the each command is and how to use it. 

## Login 

This is the exactly same as `oc login`.
So if you already logined to OpenShift(Kubernetes) cluster, then you don't need to do it again.

~~~
isv-cli login --username=joe --password=redhat
~~~
  
## Must Gather 
must-gather is for gathering debugging data for ISV service managed operator.


- Syntax
  ~~~
  isv-cli must-gather --image=${ISV-OPERATOR-NAME}-must-gather:0.0.1
  ~~~

- Tarball format
  must-gather.local.XXX/must-gather.tar
  ~~~
  isv-cli must-gather --image=quay.io/jooholee/nfs-provisioner-operator-must-gather:0.0.1
  ~~~

- Untar format 
  must-gather.local.XXX/quay.io/must-gather-image-sha256XXXX
  ~~~
  isv-cli must-gather --image=quay.io/jooholee/nfs-provisioner-operator-must-gather:0.0.1 --notar
  ~~~

- Change destination dir
  ~~~
  isv-cli must-gather --image=quay.io/jooholee/nfs-provisioner-operator-must-gather:0.0.1 --dest-dir /tmp/
  ~~~


## Logout 

This is the exactly same as `oc logout`.

- Usage
  ~~~
  isv-cli logout
  ~~~