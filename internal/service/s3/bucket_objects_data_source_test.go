package s3_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/s3"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/nij4t/terraform-provider-aws/internal/acctest"
)

func TestAccS3BucketObjectsDataSource_basic(t *testing.T) {
	rInt := sdkacctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                  func() { acctest.PreCheck(t) },
		ErrorCheck:                acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:                 acctest.Providers,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: testAccObjectsResourcesDataSourceConfig(rInt), // NOTE: contains no data source
				// Does not need Check
			},
			{
				Config: testAccObjectsBasicDataSourceConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObjectsExistsDataSource("data.aws_s3_bucket_objects.yesh"),
					resource.TestCheckResourceAttr("data.aws_s3_bucket_objects.yesh", "keys.#", "2"),
					resource.TestCheckResourceAttr("data.aws_s3_bucket_objects.yesh", "keys.0", "arch/navajo/north_window"),
					resource.TestCheckResourceAttr("data.aws_s3_bucket_objects.yesh", "keys.1", "arch/navajo/sand_dune"),
				),
			},
		},
	})
}

func TestAccS3BucketObjectsDataSource_basicViaAccessPoint(t *testing.T) {
	rInt := sdkacctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                  func() { acctest.PreCheck(t) },
		ErrorCheck:                acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:                 acctest.Providers,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: testAccObjectsResourcesPlusAccessPointDataSourceConfig(rInt), // NOTE: contains no data source
				// Does not need Check
			},
			{
				Config: testAccObjectsBasicViaAccessPointDataSourceConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObjectsExistsDataSource("data.aws_s3_bucket_objects.yesh"),
					resource.TestCheckResourceAttr("data.aws_s3_bucket_objects.yesh", "keys.#", "2"),
					resource.TestCheckResourceAttr("data.aws_s3_bucket_objects.yesh", "keys.0", "arch/navajo/north_window"),
					resource.TestCheckResourceAttr("data.aws_s3_bucket_objects.yesh", "keys.1", "arch/navajo/sand_dune"),
				),
			},
		},
	})
}

func TestAccS3BucketObjectsDataSource_all(t *testing.T) {
	rInt := sdkacctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                  func() { acctest.PreCheck(t) },
		ErrorCheck:                acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:                 acctest.Providers,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: testAccObjectsResourcesDataSourceConfig(rInt), // NOTE: contains no data source
				// Does not need Check
			},
			{
				Config: testAccObjectsAllDataSourceConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObjectsExistsDataSource("data.aws_s3_bucket_objects.yesh"),
					resource.TestCheckResourceAttr("data.aws_s3_bucket_objects.yesh", "keys.#", "7"),
					resource.TestCheckResourceAttr("data.aws_s3_bucket_objects.yesh", "keys.0", "arch/courthouse_towers/landscape"),
					resource.TestCheckResourceAttr("data.aws_s3_bucket_objects.yesh", "keys.1", "arch/navajo/north_window"),
					resource.TestCheckResourceAttr("data.aws_s3_bucket_objects.yesh", "keys.2", "arch/navajo/sand_dune"),
					resource.TestCheckResourceAttr("data.aws_s3_bucket_objects.yesh", "keys.3", "arch/partition/park_avenue"),
					resource.TestCheckResourceAttr("data.aws_s3_bucket_objects.yesh", "keys.4", "arch/rubicon"),
					resource.TestCheckResourceAttr("data.aws_s3_bucket_objects.yesh", "keys.5", "arch/three_gossips/broken"),
					resource.TestCheckResourceAttr("data.aws_s3_bucket_objects.yesh", "keys.6", "arch/three_gossips/turret"),
				),
			},
		},
	})
}

func TestAccS3BucketObjectsDataSource_prefixes(t *testing.T) {
	rInt := sdkacctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                  func() { acctest.PreCheck(t) },
		ErrorCheck:                acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:                 acctest.Providers,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: testAccObjectsResourcesDataSourceConfig(rInt), // NOTE: contains no data source
				// Does not need Check
			},
			{
				Config: testAccObjectsPrefixesDataSourceConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObjectsExistsDataSource("data.aws_s3_bucket_objects.yesh"),
					resource.TestCheckResourceAttr("data.aws_s3_bucket_objects.yesh", "keys.#", "1"),
					resource.TestCheckResourceAttr("data.aws_s3_bucket_objects.yesh", "keys.0", "arch/rubicon"),
					resource.TestCheckResourceAttr("data.aws_s3_bucket_objects.yesh", "common_prefixes.#", "4"),
					resource.TestCheckResourceAttr("data.aws_s3_bucket_objects.yesh", "common_prefixes.0", "arch/courthouse_towers/"),
					resource.TestCheckResourceAttr("data.aws_s3_bucket_objects.yesh", "common_prefixes.1", "arch/navajo/"),
					resource.TestCheckResourceAttr("data.aws_s3_bucket_objects.yesh", "common_prefixes.2", "arch/partition/"),
					resource.TestCheckResourceAttr("data.aws_s3_bucket_objects.yesh", "common_prefixes.3", "arch/three_gossips/"),
				),
			},
		},
	})
}

func TestAccS3BucketObjectsDataSource_encoded(t *testing.T) {
	rInt := sdkacctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                  func() { acctest.PreCheck(t) },
		ErrorCheck:                acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:                 acctest.Providers,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: testAccObjectsExtraResourceDataSourceConfig(rInt), // NOTE: contains no data source
				// Does not need Check
			},
			{
				Config: testAccObjectsEncodedDataSourceConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObjectsExistsDataSource("data.aws_s3_bucket_objects.yesh"),
					resource.TestCheckResourceAttr("data.aws_s3_bucket_objects.yesh", "keys.#", "2"),
					resource.TestCheckResourceAttr("data.aws_s3_bucket_objects.yesh", "keys.0", "arch/ru+b+ic+on"),
					resource.TestCheckResourceAttr("data.aws_s3_bucket_objects.yesh", "keys.1", "arch/rubicon"),
				),
			},
		},
	})
}

func TestAccS3BucketObjectsDataSource_maxKeys(t *testing.T) {
	rInt := sdkacctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                  func() { acctest.PreCheck(t) },
		ErrorCheck:                acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:                 acctest.Providers,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: testAccObjectsResourcesDataSourceConfig(rInt), // NOTE: contains no data source
				// Does not need Check
			},
			{
				Config: testAccObjectsMaxKeysDataSourceConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObjectsExistsDataSource("data.aws_s3_bucket_objects.yesh"),
					resource.TestCheckResourceAttr("data.aws_s3_bucket_objects.yesh", "keys.#", "2"),
					resource.TestCheckResourceAttr("data.aws_s3_bucket_objects.yesh", "keys.0", "arch/courthouse_towers/landscape"),
					resource.TestCheckResourceAttr("data.aws_s3_bucket_objects.yesh", "keys.1", "arch/navajo/north_window"),
				),
			},
		},
	})
}

func TestAccS3BucketObjectsDataSource_startAfter(t *testing.T) {
	rInt := sdkacctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                  func() { acctest.PreCheck(t) },
		ErrorCheck:                acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:                 acctest.Providers,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: testAccObjectsResourcesDataSourceConfig(rInt), // NOTE: contains no data source
				// Does not need Check
			},
			{
				Config: testAccObjectsStartAfterDataSourceConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObjectsExistsDataSource("data.aws_s3_bucket_objects.yesh"),
					resource.TestCheckResourceAttr("data.aws_s3_bucket_objects.yesh", "keys.#", "1"),
					resource.TestCheckResourceAttr("data.aws_s3_bucket_objects.yesh", "keys.0", "arch/three_gossips/turret"),
				),
			},
		},
	})
}

func TestAccS3BucketObjectsDataSource_fetchOwner(t *testing.T) {
	rInt := sdkacctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                  func() { acctest.PreCheck(t) },
		ErrorCheck:                acctest.ErrorCheck(t, s3.EndpointsID),
		Providers:                 acctest.Providers,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: testAccObjectsResourcesDataSourceConfig(rInt), // NOTE: contains no data source
				// Does not need Check
			},
			{
				Config: testAccObjectsOwnersDataSourceConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckObjectsExistsDataSource("data.aws_s3_bucket_objects.yesh"),
					resource.TestCheckResourceAttr("data.aws_s3_bucket_objects.yesh", "keys.#", "2"),
					resource.TestCheckResourceAttr("data.aws_s3_bucket_objects.yesh", "owners.#", "2"),
				),
			},
		},
	})
}

func testAccCheckObjectsExistsDataSource(addr string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[addr]
		if !ok {
			return fmt.Errorf("Can't find S3 objects data source: %s", addr)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("S3 objects data source ID not set")
		}

		return nil
	}
}

func testAccObjectsResourcesDataSourceConfig(randInt int) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "objects_bucket" {
  bucket = "tf-acc-objects-test-bucket-%d"
}

resource "aws_s3_bucket_object" "object1" {
  bucket  = aws_s3_bucket.objects_bucket.id
  key     = "arch/three_gossips/turret"
  content = "Delicate"
}

resource "aws_s3_bucket_object" "object2" {
  bucket  = aws_s3_bucket.objects_bucket.id
  key     = "arch/three_gossips/broken"
  content = "Dark Angel"
}

resource "aws_s3_bucket_object" "object3" {
  bucket  = aws_s3_bucket.objects_bucket.id
  key     = "arch/navajo/north_window"
  content = "Balanced Rock"
}

resource "aws_s3_bucket_object" "object4" {
  bucket  = aws_s3_bucket.objects_bucket.id
  key     = "arch/navajo/sand_dune"
  content = "Queen Victoria Rock"
}

resource "aws_s3_bucket_object" "object5" {
  bucket  = aws_s3_bucket.objects_bucket.id
  key     = "arch/partition/park_avenue"
  content = "Double-O"
}

resource "aws_s3_bucket_object" "object6" {
  bucket  = aws_s3_bucket.objects_bucket.id
  key     = "arch/courthouse_towers/landscape"
  content = "Fiery Furnace"
}

resource "aws_s3_bucket_object" "object7" {
  bucket  = aws_s3_bucket.objects_bucket.id
  key     = "arch/rubicon"
  content = "Devils Garden"
}
`, randInt)
}

func testAccObjectsResourcesPlusAccessPointDataSourceConfig(randInt int) string {
	return testAccObjectsResourcesDataSourceConfig(randInt) + fmt.Sprintf(`
resource "aws_s3_access_point" "test" {
  bucket = aws_s3_bucket.objects_bucket.bucket
  name   = "tf-objects-test-access-point-%[1]d"
}
`, randInt)
}

func testAccObjectsBasicDataSourceConfig(randInt int) string {
	return fmt.Sprintf(`
%s

data "aws_s3_bucket_objects" "yesh" {
  bucket    = aws_s3_bucket.objects_bucket.id
  prefix    = "arch/navajo/"
  delimiter = "/"
}
`, testAccObjectsResourcesDataSourceConfig(randInt))
}

func testAccObjectsBasicViaAccessPointDataSourceConfig(randInt int) string {
	return testAccObjectsResourcesPlusAccessPointDataSourceConfig(randInt) + `
data "aws_s3_bucket_objects" "yesh" {
  bucket    = aws_s3_access_point.test.arn
  prefix    = "arch/navajo/"
  delimiter = "/"
}
`
}

func testAccObjectsAllDataSourceConfig(randInt int) string {
	return fmt.Sprintf(`
%s

data "aws_s3_bucket_objects" "yesh" {
  bucket = aws_s3_bucket.objects_bucket.id
}
`, testAccObjectsResourcesDataSourceConfig(randInt))
}

func testAccObjectsPrefixesDataSourceConfig(randInt int) string {
	return fmt.Sprintf(`
%s

data "aws_s3_bucket_objects" "yesh" {
  bucket    = aws_s3_bucket.objects_bucket.id
  prefix    = "arch/"
  delimiter = "/"
}
`, testAccObjectsResourcesDataSourceConfig(randInt))
}

func testAccObjectsExtraResourceDataSourceConfig(randInt int) string {
	return fmt.Sprintf(`
%s

resource "aws_s3_bucket_object" "object8" {
  bucket  = aws_s3_bucket.objects_bucket.id
  key     = "arch/ru b ic on"
  content = "Goose Island"
}
`, testAccObjectsResourcesDataSourceConfig(randInt))
}

func testAccObjectsEncodedDataSourceConfig(randInt int) string {
	return fmt.Sprintf(`
%s

data "aws_s3_bucket_objects" "yesh" {
  bucket        = aws_s3_bucket.objects_bucket.id
  encoding_type = "url"
  prefix        = "arch/ru"
}
`, testAccObjectsExtraResourceDataSourceConfig(randInt))
}

func testAccObjectsMaxKeysDataSourceConfig(randInt int) string {
	return fmt.Sprintf(`
%s

data "aws_s3_bucket_objects" "yesh" {
  bucket   = aws_s3_bucket.objects_bucket.id
  max_keys = 2
}
`, testAccObjectsResourcesDataSourceConfig(randInt))
}

func testAccObjectsStartAfterDataSourceConfig(randInt int) string {
	return fmt.Sprintf(`
%s

data "aws_s3_bucket_objects" "yesh" {
  bucket      = aws_s3_bucket.objects_bucket.id
  start_after = "arch/three_gossips/broken"
}
`, testAccObjectsResourcesDataSourceConfig(randInt))
}

func testAccObjectsOwnersDataSourceConfig(randInt int) string {
	return fmt.Sprintf(`
%s

data "aws_s3_bucket_objects" "yesh" {
  bucket      = aws_s3_bucket.objects_bucket.id
  prefix      = "arch/three_gossips/"
  fetch_owner = true
}
`, testAccObjectsResourcesDataSourceConfig(randInt))
}
