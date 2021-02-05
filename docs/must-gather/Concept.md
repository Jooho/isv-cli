# ISV Must Gather Concept

[Must Gater](https://github.com/openshift/must-gather) is a tool for collecting openshift cluster data. 
In order to debug ISV images when it has issues, we will use this tool especially using a must-gather image. 



## Requirement Analysis
- ISV must provide must-gather implementation.
- ISV must provide functest implementation.
  - Expected files must be listed up
- The user must have admin role to gather data.


## OpenShift Must Gather Analysis
- Must Gather need a cluster-admin role to gather data.
  > oc adm must-gather

- Must Gather can add plug-ins image for each component.
  > oc adm must-gather --image=quay.io/ocs-dev/ocs-must-gather

- Must Gather can execute a specific must-gather script independently.
  > oc adm must-gather -- gather_etcd

- Using `oc adm must-gather`, it copies all data under /must-gather folder in the container to host directory
    - `--dest-dir` can change the host directory where the data will be copied
  - A client need to tar the folder for uploading
- After gathering data, a client should archive the data into tar 



### Links
- [Must-gather Repo](https://github.com/openshift/must-gather)
- [Must-gather White Pager](https://github.com/openshift/enhancements/blob/master/enhancements/oc/must-gather.md)

### Examples ###
- [local-storage-operator](https://github.com/openshift/local-storage-operator/blob/master/must-gather/gather) - bash shell 
- [OCS operator](https://github.com/openshift/ocs-operator/tree/master/must-gather)


## Idea for ISV must-gather

### Workflow

openshift/release  <--------> ISV-Integate-Repo <---------> ISV Operater Repo

openshift/release
- Prow Job
- Prow Conf

ISV-Integate-Repo
- Test Conf
  - Operator Image (=> operator test)
  - Operator Bundle Image (=> olm test)
  - Operator Index Image (=> olm test)
  - Operator Must Gather Image (=> must-gather test)

### Steps
1. ISV develop an operator image/bundle/index.
2. ISV uploads them to Red Hat to get a certificate.
   a. If succeed, ISV send PR(update image version) to ISV-Integate-Repo for triggering test(must-gather/e2e)
   b. If failed, back to step 1
3. PR show test jobs.
   a. If all test jobs succeed, Red Hat confirms that the image has no issue running on OSD and the must-gather image is also gathering data correctly.
   b. If some jobs failed, ISV must recheck their images. Red Hat needs to provide debugging data(TBD- maybe dashboard log)



Individual gather file can be executable.
gather-debug.log should have each command status.

**TODO**
- NFS Provisioner must-gather 만들기
- Standard ISV must-gather 구조 잡기
- ISV-Integate-Repo conf 구조 생각해보기
- e2e test conf 구조 잡기





