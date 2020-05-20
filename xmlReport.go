/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/getgauge/xml-report/logger"

	"github.com/getgauge/common"
	"github.com/getgauge/xml-report/builder"
	"github.com/getgauge/xml-report/gauge_messages"
)

const (
	defaultReportsDir           = "reports"
	gaugeReportsDirEnvName      = "gauge_reports_dir" // directory where reports are generated by plugins
	executionAction             = "execution"
	pluginActionEnv             = "xml-report_action"
	xmlReport                   = "xml-report"
	overwriteReportsEnvProperty = "overwrite_reports"
	resultFile                  = "result.xml"
	timeFormat                  = "2006-01-02 15.04.05"
)

var projectRoot string
var pluginDir string

func createReport(suiteResult *gauge_messages.SuiteExecutionResult) {
	dir := createReportsDirectory()
	bytes, err := builder.NewXmlBuilder(0).GetXmlContent(suiteResult)
	if err != nil {
		logger.Fatal("Report generation failed: %s \n", err)
	}
	err = writeResultFile(dir, bytes)
	if err != nil {
		logger.Fatal("Report generation failed: %s \n", err)
	}
	logger.Info("Successfully generated xml-report to => %s\n", dir)
}

func writeResultFile(reportDir string, bytes []byte) error {
	resultPath := filepath.Join(reportDir, resultFile)
	err := ioutil.WriteFile(resultPath, bytes, common.NewFilePermissions)
	if err != nil {
		return fmt.Errorf("failed to copy file: %s %s\n ", resultFile, err)
	}
	return nil
}

func findPluginAndProjectRoot() {
	projectRoot = os.Getenv(common.GaugeProjectRootEnv)
	if projectRoot == "" {
		logger.Fatal("Environment variable '%s' is not set. \n", common.GaugeProjectRootEnv)
	}
	var err error
	pluginDir, err = os.Getwd()
	if err != nil {
		logger.Fatal("Error finding current working directory: %s \n", err)
	}
}

func createReportsDirectory() string {
	reportsDir, err := filepath.Abs(os.Getenv(gaugeReportsDirEnvName))
	if reportsDir == "" || err != nil {
		reportsDir = defaultReportsDir
	}
	currentReportDir := filepath.Join(reportsDir, xmlReport, getNameGen().randomName())
	createDirectory(currentReportDir)
	return currentReportDir
}

func createDirectory(dir string) {
	if common.DirExists(dir) {
		return
	}
	if err := os.MkdirAll(dir, common.NewDirectoryPermissions); err != nil {
		logger.Fatal("Failed to create directory %s: %s\n", defaultReportsDir, err)
	}
}

func getNameGen() nameGenerator {
	if shouldOverwriteReports() {
		return emptyNameGenerator{}
	}
	return timeStampedNameGenerator{}
}

type nameGenerator interface {
	randomName() string
}
type timeStampedNameGenerator struct{}

func (T timeStampedNameGenerator) randomName() string {
	return time.Now().Format(timeFormat)
}

type emptyNameGenerator struct{}

func (T emptyNameGenerator) randomName() string {
	return ""
}

func shouldOverwriteReports() bool {
	envValue := os.Getenv(overwriteReportsEnvProperty)
	if strings.ToLower(envValue) == "true" {
		return true
	}
	return false
}
