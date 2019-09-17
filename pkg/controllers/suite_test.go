// SPDX-License-Identifier: MIT

package controllers

import (
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"

	ginkgoconfig "github.com/onsi/ginkgo/config"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	// +kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var cfg *rest.Config
var k8sClient client.Client
var testEnv *envtest.Environment

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	/* RunSpecsWithDefaultAndCustomReporters(t,
	"Controller Suite",
	[]Reporter{envtest.NewlineReporter{}}) */

	resultdir, present := os.LookupEnv("TEST_RESULT_DIR")
	if !present {
		resultdir = "build/test-results"
	}
	_ = os.MkdirAll(resultdir, os.ModePerm)
	junitReporter := reporters.NewJUnitReporter(filepath.Join(resultdir, "tests-junit-report.xml"))
	ginkgoconfig.DefaultReporterConfig.Verbose = true
	RunSpecsWithDefaultAndCustomReporters(t, "Controller Suite", []Reporter{junitReporter})
}

// var _ = BeforeSuite(func(done Done) {
// 	logf.SetLogger(zap.LoggerTo(GinkgoWriter, true))

// 	By("bootstrapping test environment")
// 	testEnv = &envtest.Environment{
// 		CRDDirectoryPaths: []string{filepath.Join("..", "config", "crd", "bases")},
// 	}

// 	cfg, err := testEnv.Start()
// 	Expect(err).ToNot(HaveOccurred())
// 	Expect(cfg).ToNot(BeNil())

// 	err = corev1.AddToScheme(scheme.Scheme)
// 	Expect(err).NotTo(HaveOccurred())

// 	// +kubebuilder:scaffold:scheme

// 	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
// 	Expect(err).ToNot(HaveOccurred())
// 	Expect(k8sClient).ToNot(BeNil())

// 	close(done)
// }, 60)

// var _ = AfterSuite(func() {
// 	By("tearing down the test environment")
// 	err := testEnv.Stop()
// 	Expect(err).ToNot(HaveOccurred())
// })
