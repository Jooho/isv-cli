# ISV Must Gather Concept

[Must Gater](https://github.com/openshift/must-gather) is a tool for collecting openshift cluster data. 
In order to debug ISV images when it has issues, we will use this tool especially using an must-gather image. 



## Requirement Analysis
- ISV must provide must-gather implimentation.
  - must-gather data should be archieved by a customer.
- ISV must provide functest implimentation.
  - Expected files must be listed up
  - Not expected empty files must be listed up
- Red Hat OSD team can gather all nessasary data executing `oc adm must-gather --image=`
- The user must have cluster-admin role to gather data.


## OpenShift Must Gather Analysis
- Must Gather need a cluster-admin role to gather data.
  > oc adm must-gather

- Must Gather can add plug-ins image for each component.
  > oc adm must-gather --image=quay.io/ocs-dev/ocs-must-gather

- Must Gather can execute a specific must-gather script independently.
  > oc adm must-gather -- gather_etcd

- Using `oc adm must-gather`, it copies all data under /must-gather folder in the container to host directory
    - `--dest-dir` can change host directory where the data will be copied
  - A client need to tar the folder for uploading
- After gathering data, a client should archieve the data into tar 



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
2. ISV upload them to Red Hat to get certificate.
   a. If succeed, ISV send PR(update image version) to ISV-Integate-Repo for triggering test(must-gather/e2e)
   b. If failed, back to step 1
3. PR show test jobs.
   a. If all test jobs succeed, Red Hat confirm that the image have no issue running on OSD an the must-gather image is also gathering data correctly.
   b. If some jobs failed, ISV must recheck their images. Red Hat need to provide debugging data(TBD- maybe dashboard log)


gather file은 독립적으로 실행이 가능해야한다.
gather-debug.log를 통해 각 status를 update한다.

**TODO**
- NFS Provisioner must-gather 만들기
- Standard ISV must-gather 구조 잡기
- ISV-Integate-Repo conf 구조 생각해보기
- e2e test conf 구조 잡기

**Question**
- e2e test will build an operator in a CI process and will do unit-test and e2e test. For ISV, from my understanding, Red Hat will use the operator image that passed Red Hat Certificate process. is it correct?
  - How can prow conf support it?

- who will trigger to gather data?

shellcheck.sh

### 가이드


#### Test 방법


# Output Structure
#

# must-gather.local.RAND/quay-io-isv-must-gather-sha256-XXXX/
# - timestamp
# - version
# - namespace-scoped-resources/ 
#   - pods/
#     - oc_get_wide.txt
#     - %POD_NAME%_oc_describe.txt
#     - %POD_NAME%.yaml
#     - %POD_NAME%.log
#   - namespace/
#     - namespace.yaml
#   - deployment/
#      - %DEPLOYMENT_NAME%.yaml
#   - deploymentConfig/
#      - %DEPLOYMENTCONFIG_NAME%.yaml
#   - replicaset/
#      - %REPLICASET_NAME%.yaml
#   - service/
#      - %SERVICE_NAME%.yaml
#   - route/
#      - %ROUTE_NAME%.yaml
#   - persistentvolumeciaim/
#      - %PVC_NAME%.yaml
#   - event/
#      - oc_get_event
# - operand/
#   - Freeform





