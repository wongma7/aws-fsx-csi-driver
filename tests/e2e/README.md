# Prerequisites
- kubernetes 1.14+ cluster whose workers (preferably 2 or more) can mount Amazon FSx for Lustre file systems

# Run
```sh
dep ensure
GO111MODULE=off go test -v -timeout 0 ./... -kubeconfig=$HOME/.kube/config -ginkgo.focus="\[fsx-csi\]" -ginkgo.skip="\[Disruptive\]" \
  -subnet-id=subnet-43c2d319 \
  -security-groups=sg-d2754e9f
```
# FAQ
- empty `go.mod` because: https://github.com/golang/go/wiki/Modules#can-an-additional-gomod-exclude-unnecessary-content-do-modules-have-the-equivalent-of-a-gitignore-file
- kubernetes uses go modules, but importing the e2e framework is hard, so use vendoring via dep
