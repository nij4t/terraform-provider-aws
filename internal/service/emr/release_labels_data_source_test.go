package emr_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/emr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/nij4t/terraform-provider-aws/internal/acctest"
)

func TestAccEMRReleaseLabels_basic(t *testing.T) {
	dataSourceResourceName := "data.aws_emr_release_labels.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, emr.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccReleaseLabelsDataSourceConfigBasic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceResourceName, "release_labels.#"),
				),
			},
		},
	})
}

func TestAccEMRReleaseLabels_prefix(t *testing.T) {
	dataSourceResourceName := "data.aws_emr_release_labels.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, emr.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccReleaseLabelsDataSourceConfigPrefix(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceResourceName, "release_labels.#"),
				),
			},
		},
	})
}

func TestAccEMRReleaseLabels_application(t *testing.T) {
	dataSourceResourceName := "data.aws_emr_release_labels.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, emr.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccReleaseLabelsDataSourceConfigApplication(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceResourceName, "release_labels.#"),
				),
			},
		},
	})
}

func TestAccEMRReleaseLabels_full(t *testing.T) {
	dataSourceResourceName := "data.aws_emr_release_labels.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, emr.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      nil,
		Steps: []resource.TestStep{
			{
				Config: testAccReleaseLabelsDataSourceConfigFull(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceResourceName, "release_labels.#"),
				),
			},
		},
	})
}

func testAccReleaseLabelsDataSourceConfigBasic() string {
	return `
data "aws_emr_release_labels" "test" {}
`
}

func testAccReleaseLabelsDataSourceConfigPrefix() string {
	return `
data "aws_emr_release_labels" "test" {
  filters {
    prefix = "emr-6"
  }
}
`
}

func testAccReleaseLabelsDataSourceConfigApplication() string {
	return `
data "aws_emr_release_labels" "test" {
  filters {
    application = "Spark@3.1.2"
  }
}
`
}

func testAccReleaseLabelsDataSourceConfigFull() string {
	return `
data "aws_emr_release_labels" "test" {
  filters {
    application = "Spark@3.1.2"
    prefix      = "emr-6"
  }
}
`
}
