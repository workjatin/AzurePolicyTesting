package policyTester

import (
	"context"   //Package context defines the Context typewhich carries deadlines, cancellation signals, and other request-scoped values across API boundaries and between processes.
	"fmt"       //Package fmt implements formatted I/O with functions analogous to C's printf and scanf
	"io/ioutil" //Package ioutil implements some I/O utility functions, same functionality is now provided by package io or package os
	"log"
	"os"
	"path/filepath" //Package filepath implements utility routines for manipulating filename paths in a way compatible with the target operating system-defined file paths.
	"testing"

	//"github.com/goproxy0/go/src/testing/internal/testdeps"
	version "github.com/hashicorp/go-version"
	install "github.com/hashicorp/hc-install"
	"github.com/hashicorp/hc-install/fs"
	"github.com/hashicorp/hc-install/product"
	"github.com/hashicorp/hc-install/releases"
	"github.com/hashicorp/hc-install/src"
	"gopkg.in/yaml.v2" // helps easily encode and decode YAML values
)

func RunGoTest(configPath string) (int, error) {

	//called in main, the root finction to start all test cases, folder contaning test cases are passes

	// make a tmp dirrctory to install terrafrom in
	tmpDir, err := ioutil.TempDir("", "tfinstall")
	if err != nil {
		return 0, fmt.Errorf("create temp dir for TF installation: %w", err)
	}
	// Differ key word makes the next finction exicute in the end
	defer os.RemoveAll(tmpDir)

	// insatall terraform in the local temp dirrectory tmpdir
	i := install.NewInstaller()
	v1, err := version.NewVersion("1.2")
	if err != nil {
		return 0, fmt.Errorf("error finding version: %w", err)
	}
	v0_14_0 := version.Must(v1,err)
	tfExecPath, err := i.Ensure(context.Background(), []src.Source{
		&fs.ExactVersion{
			Product: product.Terraform,
			Version: v0_14_0,
		},
		&releases.ExactVersion{
			Product: product.Terraform,
			Version: v0_14_0,
		},
	})
	if err != nil {
		return 0, fmt.Errorf("locate Terraform binary: %w", err)
	}

	//returns a list of file info from the config path(Test folder)
	files, err := ioutil.ReadDir(configPath)
	if err != nil {
		return 0, fmt.Errorf("could not read directory path %s: %w", configPath, err)
	}

	//create a slice of pointer type Struct testRunner
	runners := make([]*testRunner, 0)

	for _, file := range files { //for all files in config path(Test)
		if filepath.Ext(file.Name()) == ".yaml" { //check if extention is yaml
			yamlFile, err := os.Open(filepath.Join(configPath, file.Name())) //open .yaml to read
			if err != nil {
				continue
			}
			
			defer yamlFile.Close() //close file once all other statements exicute in the block
			
			byteValue, _ := ioutil.ReadAll(yamlFile)
			
			var testConfig TestConfig                                      // making a var to store the input forom the .yaml file
			if err := yaml.Unmarshal(byteValue, &testConfig); err != nil { // unpack the ymal file into the testConfig
				log.Printf("Could not unmarshal file %s: %v", file.Name(), err)
				continue
			}

			runners = append(runners, newTestRunner(tfExecPath, configPath, testConfig)) // adds current test config to runners list
			//fmt.Printf("config\n%+v\n",testConfig)
		}
	}



	//make slice of type nternalTest
	// type InternalTest struct {
	// 	Name string
	// 	F    func(*T)
	// }
	tests := make([]testing.InternalTest, 0)

	for _, runner := range runners {
		tests = append(tests, testing.InternalTest{
			Name: runner.config.Name, //contains name of the test case
			F:    runner.Test,        //points to the test function of runner
		})
	}

	
	//runs all tests cases without passing go test commanad
	t := new(TestDeps)
	return testing.MainStart(t, tests, []testing.InternalBenchmark{}, []testing.InternalFuzzTarget{}, []testing.InternalExample{}).Run(), nil
}
