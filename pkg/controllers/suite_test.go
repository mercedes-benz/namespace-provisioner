// SPDX-License-Identifier: MIT

package controllers

import (
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/reporters"
	. "github.com/onsi/gomega"

	"github.com/onsi/ginkgo/config"
)

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)
	resultDir := os.Getenv("TEST_RESULT_DIR")
	if resultDir == "" {
		resultDir = "../../build/test-results"
	}
	if err := os.MkdirAll(resultDir, os.ModePerm); err != nil {
		panic(err)
	}
	junitReporter := reporters.NewJUnitReporter(filepath.Join(resultDir, "tests-junit-report.xml"))
	config.DefaultReporterConfig.Verbose = true
	RunSpecsWithDefaultAndCustomReporters(t, "Controllers Suite", []Reporter{junitReporter})
}
