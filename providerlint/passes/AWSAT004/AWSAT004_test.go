package AWSAT004_test

import (
	"testing"

	_ "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/nij4t/terraform-provider-aws/providerlint/passes/AWSAT004"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAWSAT004(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, AWSAT004.Analyzer, "a")
}
