# Test commands
Test image is [here](https://github.com/Jooho/isv-smoke-must-gather). There are 5 different tags to test each objects but the release version must be the same as isv-cli release version.

- Standard test cmd
  ~~~
  go run ./cmd/isv-cli.go must-gather --image quay.io/jooholee/isv-smoke-must-gather:event  --dest-dir ~/dev/Managed_Git/operator-projects/tmp/cls
  ~~~
- Browser feature test
  ~~~
  go run ./cmd/isv-cli.go must-gather --image quay.io/jooholee/isv-smoke-must-gather:event --browser --dest-dir ~/dev/Managed_Git/operator-projects/tmp/cls
  ~~~