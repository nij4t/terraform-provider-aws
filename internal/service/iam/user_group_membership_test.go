package iam_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/nij4t/terraform-provider-aws/internal/acctest"
	"github.com/nij4t/terraform-provider-aws/internal/conns"
)

func TestAccIAMUserGroupMembership_basic(t *testing.T) {
	rString := sdkacctest.RandString(8)
	userName1 := fmt.Sprintf("tf-acc-ugm-basic-user1-%s", rString)
	userName2 := fmt.Sprintf("tf-acc-ugm-basic-user2-%s", rString)
	groupName1 := fmt.Sprintf("tf-acc-ugm-basic-group1-%s", rString)
	groupName2 := fmt.Sprintf("tf-acc-ugm-basic-group2-%s", rString)
	groupName3 := fmt.Sprintf("tf-acc-ugm-basic-group3-%s", rString)

	usersAndGroupsConfig := testAccUserGroupMembershipUsersAndGroupsConfig(userName1, userName2, groupName1, groupName2, groupName3)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, iam.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccUserGroupMembershipDestroy,
		Steps: []resource.TestStep{
			// simplest test
			{
				Config: usersAndGroupsConfig + testAccUserGroupMembershipInitConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aws_iam_user_group_membership.user1_test1", "user", userName1),
					testAccUserGroupMembershipCheckGroupListForUser(userName1, []string{groupName1}, []string{groupName2, groupName3}),
				),
			},
			{
				ResourceName:      "aws_iam_user_group_membership.user1_test1",
				ImportState:       true,
				ImportStateIdFunc: testAccUserGroupMembershipImportStateIdFunc("aws_iam_user_group_membership.user1_test1"),
				// We do not have a way to align IDs since the Create function uses resource.UniqueId()
				// Failed state verification, resource with ID USER/GROUP not found
				//ImportStateVerify: true,
				ImportStateCheck: func(s []*terraform.InstanceState) error {
					if len(s) != 1 {
						return fmt.Errorf("expected 1 state: %#v", s)
					}

					return nil
				},
			},
			// test adding an additional group to an existing resource
			{
				Config: usersAndGroupsConfig + testAccUserGroupMembershipAddOneConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aws_iam_user_group_membership.user1_test1", "user", userName1),
					testAccUserGroupMembershipCheckGroupListForUser(userName1, []string{groupName1, groupName2}, []string{groupName3}),
				),
			},
			// test adding multiple resources for the same user, and resources with the same groups for another user
			{
				Config: usersAndGroupsConfig + testAccUserGroupMembershipAddAllConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aws_iam_user_group_membership.user1_test1", "user", userName1),
					resource.TestCheckResourceAttr("aws_iam_user_group_membership.user1_test2", "user", userName1),
					resource.TestCheckResourceAttr("aws_iam_user_group_membership.user2_test1", "user", userName2),
					resource.TestCheckResourceAttr("aws_iam_user_group_membership.user2_test2", "user", userName2),
					testAccUserGroupMembershipCheckGroupListForUser(userName1, []string{groupName1, groupName2, groupName3}, []string{}),
					testAccUserGroupMembershipCheckGroupListForUser(userName2, []string{groupName1, groupName2, groupName3}, []string{}),
				),
			},
			// test that nothing happens when we apply the same config again
			{
				Config: usersAndGroupsConfig + testAccUserGroupMembershipAddAllConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aws_iam_user_group_membership.user1_test1", "user", userName1),
					resource.TestCheckResourceAttr("aws_iam_user_group_membership.user1_test2", "user", userName1),
					resource.TestCheckResourceAttr("aws_iam_user_group_membership.user2_test1", "user", userName2),
					resource.TestCheckResourceAttr("aws_iam_user_group_membership.user2_test2", "user", userName2),
					testAccUserGroupMembershipCheckGroupListForUser(userName1, []string{groupName1, groupName2, groupName3}, []string{}),
					testAccUserGroupMembershipCheckGroupListForUser(userName2, []string{groupName1, groupName2, groupName3}, []string{}),
				),
			},
			// test removing a group
			{
				Config: usersAndGroupsConfig + testAccUserGroupMembershipRemoveGroupConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aws_iam_user_group_membership.user1_test1", "user", userName1),
					resource.TestCheckResourceAttr("aws_iam_user_group_membership.user1_test2", "user", userName1),
					resource.TestCheckResourceAttr("aws_iam_user_group_membership.user2_test1", "user", userName2),
					resource.TestCheckResourceAttr("aws_iam_user_group_membership.user2_test2", "user", userName2),
					testAccUserGroupMembershipCheckGroupListForUser(userName1, []string{groupName1, groupName3}, []string{groupName2}),
					testAccUserGroupMembershipCheckGroupListForUser(userName2, []string{groupName1, groupName2}, []string{groupName3}),
				),
			},
			// test removing a resource
			{
				Config: usersAndGroupsConfig + testAccUserGroupMembershipDeleteResourceConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aws_iam_user_group_membership.user1_test1", "user", userName1),
					resource.TestCheckResourceAttr("aws_iam_user_group_membership.user1_test2", "user", userName1),
					resource.TestCheckResourceAttr("aws_iam_user_group_membership.user2_test1", "user", userName2),
					testAccUserGroupMembershipCheckGroupListForUser(userName1, []string{groupName1, groupName3}, []string{groupName2}),
					testAccUserGroupMembershipCheckGroupListForUser(userName2, []string{groupName1}, []string{groupName2, groupName3}),
				),
			},
		},
	})
}

func testAccUserGroupMembershipDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).IAMConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type == "aws_iam_user_group_membership" {
			input := &iam.ListGroupsForUserInput{
				UserName: aws.String(rs.Primary.Attributes["user"]),
			}
			foundGroups := 0
			err := conn.ListGroupsForUserPages(input, func(page *iam.ListGroupsForUserOutput, lastPage bool) bool {
				if len(page.Groups) > 0 {
					foundGroups = foundGroups + len(page.Groups)
				}
				return !lastPage
			})
			if err != nil {
				if tfawserr.ErrMessageContains(err, iam.ErrCodeNoSuchEntityException, "") {
					continue
				}
				return err
			}
			if foundGroups > 0 {
				return fmt.Errorf("Expected all group membership for user to be removed, found: %d", foundGroups)
			}
		}
	}

	return nil
}

func testAccUserGroupMembershipCheckGroupListForUser(userName string, groups []string, groupsNeg []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMConn

		// get list of groups for user
		userGroupList, err := conn.ListGroupsForUser(&iam.ListGroupsForUserInput{
			UserName: &userName,
		})
		if err != nil {
			return fmt.Errorf("Error validing user group list for %s: %s", userName, err)
		}

		// check required groups
	GROUP:
		for _, group := range groups {
			for _, groupFound := range userGroupList.Groups {
				if group == *groupFound.GroupName {
					continue GROUP // found our group, start checking the next one
				}
			}
			// group not found, return an error
			return fmt.Errorf("Required group not found for %s: %s", userName, group)
		}

		// check that none of groupsNeg are present
		for _, group := range groupsNeg {
			for _, groupFound := range userGroupList.Groups {
				if group == *groupFound.GroupName {
					return fmt.Errorf("Unexpected group found for %s: %s", userName, group)
				}
			}
		}

		return nil
	}
}

func testAccUserGroupMembershipImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		groupCount, _ := strconv.Atoi(rs.Primary.Attributes["groups.#"])
		stateId := rs.Primary.Attributes["user"]
		for i := 0; i < groupCount; i++ {
			groupName := rs.Primary.Attributes[fmt.Sprintf("group.%d", i)]
			stateId = fmt.Sprintf("%s/%s", stateId, groupName)
		}
		return stateId, nil
	}
}

// users and groups for all other tests
func testAccUserGroupMembershipUsersAndGroupsConfig(userName1, userName2, groupName1, groupName2, groupName3 string) string {
	return fmt.Sprintf(`
resource "aws_iam_user" "user1" {
  name          = "%s"
  force_destroy = true
}

resource "aws_iam_user" "user2" {
  name          = "%s"
  force_destroy = true
}

resource "aws_iam_group" "group1" {
  name = "%s"
}

resource "aws_iam_group" "group2" {
  name = "%s"
}

resource "aws_iam_group" "group3" {
  name = "%s"
}
`, userName1, userName2, groupName1, groupName2, groupName3)
}

// associate users and groups
const testAccUserGroupMembershipInitConfig = `
resource "aws_iam_user_group_membership" "user1_test1" {
  user = aws_iam_user.user1.name
  groups = [
    aws_iam_group.group1.name,
  ]
}
`

const testAccUserGroupMembershipAddOneConfig = `
resource "aws_iam_user_group_membership" "user1_test1" {
  user = aws_iam_user.user1.name
  groups = [
    aws_iam_group.group1.name,
    aws_iam_group.group2.name,
  ]
}
`

const testAccUserGroupMembershipAddAllConfig = `
resource "aws_iam_user_group_membership" "user1_test1" {
  user = aws_iam_user.user1.name
  groups = [
    aws_iam_group.group1.name,
    aws_iam_group.group2.name,
  ]
}

resource "aws_iam_user_group_membership" "user1_test2" {
  user = aws_iam_user.user1.name
  groups = [
    aws_iam_group.group3.name,
  ]
}

resource "aws_iam_user_group_membership" "user2_test1" {
  user = aws_iam_user.user2.name
  groups = [
    aws_iam_group.group1.name,
  ]
}

resource "aws_iam_user_group_membership" "user2_test2" {
  user = aws_iam_user.user2.name
  groups = [
    aws_iam_group.group2.name,
    aws_iam_group.group3.name,
  ]
}
`

// test removing a group
const testAccUserGroupMembershipRemoveGroupConfig = `
resource "aws_iam_user_group_membership" "user1_test1" {
  user = aws_iam_user.user1.name
  groups = [
    aws_iam_group.group1.name,
  ]
}

resource "aws_iam_user_group_membership" "user1_test2" {
  user = aws_iam_user.user1.name
  groups = [
    aws_iam_group.group3.name,
  ]
}

resource "aws_iam_user_group_membership" "user2_test1" {
  user = aws_iam_user.user2.name
  groups = [
    aws_iam_group.group1.name,
  ]
}

resource "aws_iam_user_group_membership" "user2_test2" {
  user = aws_iam_user.user2.name
  groups = [
    aws_iam_group.group2.name,
  ]
}
`

// test deleting an entity
const testAccUserGroupMembershipDeleteResourceConfig = `
resource "aws_iam_user_group_membership" "user1_test1" {
  user = aws_iam_user.user1.name
  groups = [
    aws_iam_group.group1.name,
  ]
}

resource "aws_iam_user_group_membership" "user1_test2" {
  user = aws_iam_user.user1.name
  groups = [
    aws_iam_group.group3.name,
  ]
}

resource "aws_iam_user_group_membership" "user2_test1" {
  user = aws_iam_user.user2.name
  groups = [
    aws_iam_group.group1.name,
  ]
}
`
