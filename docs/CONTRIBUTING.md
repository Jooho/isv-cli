# CONTRIBUTE

## Version Change
~~~
vi Makefile
CLI_VERSION ?= v0.2-alpha   <== update
~~~
## Build isv-cli
~~~
make build
~~~

## Build and push isv-cli image
~~~
make build
make cli-image
~~~

## Download a specific isv-cli release, build and push isv-cli image
  ~~~
  # Change Version

  make clean
  make cli-image
  ~~~
## Test
~~~
make test
~~~

## Clean repo before commit
~~~
make clean
~~~