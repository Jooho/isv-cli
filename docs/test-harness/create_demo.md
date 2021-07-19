# Demo Senario.

## Situation

Operator: RHODS operator
ADDON: RHODS
Dependent ADDONs: None

## Prerequisites
~~~
git clone https://github.com/Jooho/isv-cli.git

make build
~~~

## Steps

**Update example-config.ini**
~~~
curl -L https://raw.githubusercontent.com/Jooho/isv-cli/main/templates/test-harness/example-config.ini -o /tmp/rhods-config.ini

vi config.ini

PRODUCT_NAME=rhods
TEST_NAMESPACE=redhat-ods-applications
~~~

**Create Test Harness Repositories**
~~~
./isv-cli test-harness create --config-path=/tmp/rhods-config.ini --dest-dir=/tmp/rhods-demo
** Create Test Harness Repositories **
** Update Variable in Test Harness Repositories **

** Test Harness Repositories are Ready: /tmp/rhods-demo **
~~~

### Manifest
**Update test script for ISV operator**
  ~~~
  cd /tmp/rhods-demo/rhods-operator-manifests

  git init 
  git remote add origin https://github.com/Jooho/rhods-operator-manifests

  vi basictests/operator.sh 

  ## Remove
  # MUST UPDATE label (deploymentconfig=jupyterhub) and running pods count (1)
  %ERROR EXPECTED%

  git add .;git commit -m "initial update";git push
  ~~~

**Test Manifests**
  ~~~
  make build run

  # If it works, push the image

  make push-image
  ~~~

### Test Harness 

**Job Test**
  ~~~
  cd ../rhods-operator-test-harness

  make job-test

  oc logs job/manifests-test-job -f

  # If it works, clean job test and push image
  make job-test-clean image
  ~~~

**Cluster Test**
  ~~~
  make cluster-test
  oc logs rhods-operator-test-harness-pod -f -c operator
  
  #If it works, clean cluster test and push src
  make cluster-test-clean

  git init 
  git remote add origin https://github.com/Jooho/rhods-operator-test-harness
  git add .;git commit -m "initial update";git push
  ~~~

