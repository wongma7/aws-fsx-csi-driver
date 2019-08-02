/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package e2e

import (
	"flag"
	"fmt"
	"log"
	"testing"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/kubernetes/test/e2e/framework"
	"k8s.io/kubernetes/test/e2e/framework/testfiles"
	"k8s.io/kubernetes/test/e2e/storage/testpatterns"
	"k8s.io/kubernetes/test/e2e/storage/testsuites"
)

func init() {
	framework.HandleFlags()
	framework.AfterReadingAllFlags(&framework.TestContext)
	// PWD is test/e2e inside the git repo
	testfiles.AddFileSource(testfiles.RootFileSource{Root: "../.."})
}

var subnetID = flag.String("subnet-id", "", "required. subnet ID in which to create FSx file systems")
var securityGroups = flag.String("security-groups", "", "required. security groups to associate FSx file systems with (comma-delimited)")

func TestFSxCSI(t *testing.T) {
	if *subnetID == "" {
		log.Fatalf("subnet-id required. subnet ID in which to create FSx file systems")
	}
	if *securityGroups == "" {
		log.Fatalf("security-groups required. security groups to associate FSx file systems with (comma-delimited)")
	}
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "FSx CSI Suite")
}

type fsxDriver struct {
	driverInfo testsuites.DriverInfo
}

var _ testsuites.TestDriver = &fsxDriver{}

// TODO implement Inline (unless it's redundant) and PreprovisionedPV
// var _ testsuites.InlineVolumeTestDriver = &fsxDriver{}
var _ testsuites.DynamicPVTestDriver = &fsxDriver{}

func InitFSxCSIDriver() testsuites.TestDriver {
	return &fsxDriver{
		driverInfo: testsuites.DriverInfo{
			Name:                 "fsx.csi.aws.com",
			SupportedFsType:      sets.NewString(""),
			SupportedMountOption: sets.NewString("flock"),
			Capabilities: map[testsuites.Capability]bool{
				testsuites.CapPersistence: true,
				testsuites.CapExec:        true,
				testsuites.CapMultiPODs:   true,
				testsuites.CapRWX:         true,
			},
		},
	}
}

func (e *fsxDriver) GetDriverInfo() *testsuites.DriverInfo {
	return &e.driverInfo
}

func (e *fsxDriver) SkipUnsupportedTest(testpatterns.TestPattern) {}

func (e *fsxDriver) PrepareTest(f *framework.Framework) (*testsuites.PerTestConfig, func()) {
	ginkgo.By("Deploying FSx CSI driver")
	cancelPodLogs := testsuites.StartPodLogs(f)

	/* TODO deploy the driver before the test.
	cleanup, err := f.CreateFromManifests(nil, "deploy/kubernetes/manifest.yaml")
	if err != nil {
		framework.Failf("Error deploying FSx CSI driver: %v", err)
	}
	*/

	return &testsuites.PerTestConfig{
			Driver:    e,
			Prefix:    "fsx",
			Framework: f,
		}, func() {
			ginkgo.By("Cleaning up FSx CSI driver")
			/* TODO deploy the driver before the test.
			cleanup()
			*/
			cancelPodLogs()
		}
}

func (e *fsxDriver) GetDynamicProvisionStorageClass(config *testsuites.PerTestConfig, fsType string) *storagev1.StorageClass {
	ns := config.Framework.Namespace.Name
	provisioner := e.driverInfo.Name
	suffix := fmt.Sprintf("%s-sc", e.driverInfo.Name)

	parameters := map[string]string{
		"subnetId":         *subnetID,
		"securityGroupIds": *securityGroups,
	}

	return testsuites.GetStorageClass(provisioner, parameters, nil, ns, suffix)
}

func (e *fsxDriver) GetClaimSize() string {
	return "3600Gi" // this is the minimum for FSx for Lustre
}

// List of testSuites to be executed in below loop
var csiTestSuites = []func() testsuites.TestSuite{
	testsuites.InitVolumesTestSuite,
	testsuites.InitVolumeIOTestSuite,
	testsuites.InitVolumeModeTestSuite,
	testsuites.InitSubPathTestSuite,
	testsuites.InitProvisioningTestSuite,
	//testsuites.InitSnapshottableTestSuite,
	testsuites.InitMultiVolumeTestSuite,
}

var _ = ginkgo.Describe("[fsx-csi] FSx CSI", func() {
	driver := InitFSxCSIDriver()
	ginkgo.Context(testsuites.GetDriverNameWithFeatureTags(InitFSxCSIDriver()), func() {
		testsuites.DefineTestSuite(driver, csiTestSuites)
	})
})
