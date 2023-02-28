package tester

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	// "time"

	tfFiles "github.com/gruntwork-io/terratest/modules/files"

	"github.com/hashicorp/terraform-exec/tfexec"
	"github.com/stretchr/testify/require"
)

type testRunner struct {
	tfExecPath string
	config     TestConfig
}

func newTestRunner(tfExecPath, configPath string, config TestConfig) *testRunner {
	runner := &testRunner{
		tfExecPath: tfExecPath,
		config:     config,
	}

	runner.config.TerraformDir = path.Join(configPath, runner.config.TerraformDir)

	return runner
}

func (runner *testRunner) Test(t *testing.T) {

	//this function contains testing logic

	t.Parallel()
	ctx := context.Background()
	//fmt.Printf("tfexce path:%s", runner.tfExecPath)
	fmt.Printf("DIR %s", runner.config.TerraformDir)
	tmpDirSetup, err := tfFiles.CopyTerraformFolderToTemp(filepath.Join(runner.config.TerraformDir, "setup"), "*")
	require.NoError(t, err, "Create temp dir for setup")

	t.Cleanup(func() {
		t.Log("Tempdir cleanup main")
		os.RemoveAll(tmpDirSetup)
	})

	setup, err := tfexec.NewTerraform(tmpDirSetup, runner.tfExecPath)
	setup.SetStdout(os.Stdout)
	setup.SetStderr(os.Stderr)
	//fmt.Printf("Setup\n%+v\n",setup)
	require.NoError(t, err, "setup: new Terraform object")
	require.NoErrorf(t, setup.Init(ctx, tfexec.Upgrade(false), tfexec.Reconfigure(true)), "setup: Init command. Directory: %s", setup.WorkingDir())
	//performs cleanup after all tests are compleated
	t.Cleanup(func() {
		t.Log("destroy cleanup main")
		if err := setup.Destroy(ctx); err != nil {
			t.Logf("Destroy Setup failed: %s", err.Error())
		}
	})

	//_, err := setup.Plan(context.Background(), out)
	//t.FailNow()
	require.NoError(t, setup.Apply(ctx, tfexec.Lock(false)), "setup: error running Apply command")

	outputs, err := setup.Output(ctx)
	require.NoError(t, err, "setup: error running Output command")

	errorMessagesExpectedParts := []string{
		runner.config.ErrorMessage,
		runner.config.ErrorCode,
	}

	vars := make([]*tfexec.VarOption, 0)

	for key, output := range outputs {
		var value string
		require.NoErrorf(t, json.Unmarshal(output.Value, &value), "setup: unmarshall value of %s from the outputs", key)

		vars = append(vars, tfexec.Var(fmt.Sprintf("%s=%v", key, value)))
	}

	//time.Sleep(30 * time.Minute) // Time for the policy to be active

	for _, c := range runner.config.Cases {
		testCase := c

		t.Run(fmt.Sprint(testCase.Variables), func(t *testing.T) {
			t.Parallel()

			//make list of vars
			testCaseVars := make([]*tfexec.VarOption, 0)
			testCaseVars = append(testCaseVars, vars...)

			for _, variable := range testCase.Variables {
				testCaseVars = append(testCaseVars, tfexec.Var(fmt.Sprintf("%s=%v", variable.Key, variable.Value)))
			}

			//temp
			temp := make([]tfexec.VarOption, 0)
			for _, p := range testCaseVars {
				temp = append(temp, *p)
			}
			t.Log(temp)

			tmpDir, err := tfFiles.CopyTerraformFolderToTemp(runner.config.TerraformDir, "*")
			require.NoError(t, err, "Create temp dir for test")

			t.Cleanup(func() {
				t.Log("Tempdir cleanup")
				os.RemoveAll(tmpDir)
			})

			tf, err := tfexec.NewTerraform(tmpDir, runner.tfExecPath)
			require.NoError(t, err, "New Terraform object")
			tf.SetStdout(os.Stdout)
			tf.SetStderr(os.Stderr)

			require.NoError(t, tf.Init(ctx, tfexec.Upgrade(false), tfexec.Reconfigure(true)), "Init command")

			t.Cleanup(func() {
				t.Log("destroy cleanup")
				destroyOptions := make([]tfexec.DestroyOption, 0)
				for _, variable := range testCaseVars {
					destroyOptions = append(destroyOptions, variable)
				}
				if err := tf.Destroy(ctx, destroyOptions...); err != nil {
					t.Logf("Destroy failed: %s", err.Error())
				}
			})

			applyOptions := make([]tfexec.ApplyOption, 0)
			for _, variable := range testCaseVars {
				applyOptions = append(applyOptions, variable)
			}

			applyOptions = append(applyOptions, tfexec.Lock(false))
			err = tf.Apply(
				ctx,
				applyOptions...,
			)

			if err != nil {
				t.Log("*****************************")
				t.Log("Error in:", testCase)
				t.Log(err)
				t.Log(err.Error())
				errString := err.Error()
				if testCase.ErrorExpected {
					t.Log("Error Expected")
					matches := 0
					for _, part := range errorMessagesExpectedParts {
						if strings.Contains(strings.ToLower(errString), strings.ToLower(part)) {
							matches++
						}
					}
					require.Equalf(t, len(errorMessagesExpectedParts), matches, "deployment failed for an unexpected reason: %s", errString)
				} else {
					require.FailNow(t, "deployment failed for an unexpected reason", errString)
				}
			} else if testCase.ErrorExpected {
				require.FailNowf(t, "values should be FORBIDDEN by policy", "%s", testCase.Variables)
			}
		})
	}
}
