# Test Harness Create command

## Background
This cmd will create Test Harness and Manifests repositories. Test Harness repo will use Manifest image as a Job.

![Image](images/test_harness_1.png)


## Process after created 2 repositories

1. Manifests
First, you should take care of manifest repository.

1-1. Update operator.sh
```
vi $MANIFESTS_REPO/basictests/operator.sh

# MUST UPDATE label (deploymentconfig=jupyterhub) and running pods count (1)
%ERROR EXPECTED%

```

1-2. Push the git repo 
~~~
git init 
git remote add origin https://github.com/$ORG/$MANIFESTS-NAME
git add .;git commit -m "initial update"; git push --set-upstream origin master
~~~

1-3. Test
~~~
make build run
~~~

1-4. Push the imagegit 
~~~
make image
~~~

2. TEST Harness
2-1. Do job test (this is using the manifests image)
~~~
make job-test
~~~
2-2. Clean job test
~~~
make job-test-clean
~~~

2-3. Push the git repo
~~~
git init 
git remote add origin https://github.com/$ORG/$TEST_HARNESS_NAME
git add .;git commit -m "initial update";git push
~~~

2-4. Build/Push an image
~~~
make image
~~~

2.5. Cluster Test
~~~
make cluster-test
~~~

2.6 Clean Cluster Test
~~~
make cluster-test-clean
~~~
