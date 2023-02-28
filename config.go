package policyTester

//defines all the structs used

//test cases
type TestConfig struct {
	Name         string           `yaml:"name"`
	Cases        []PolicyTestCase `yaml:"cases"`
	//Effect       string           `yaml:"effect"`
	TerraformDir string           `yaml:"terraformDir"`
	ErrorMessage string           `yaml:"errorMessage"`
	ErrorCode    string           `yaml:"errorCode"`
}

type PolicyTestCase struct {
	ErrorExpected bool               `yaml:"errorExpected"`
	Variables     []TestCaseVariable `yaml:"variables"`
}

type TestCaseVariable struct {
	Key   string      `yaml:"key"`
	Value interface{} `yaml:"value"`
}
