package namevaluesfilters

import (
	"testing"

	"github.com/nij4t/terraform-provider-aws/internal/acctest"
)

func TestNameValuesFiltersEc2Tags(t *testing.T) {
	testCases := []struct {
		name    string
		filters NameValuesFilters
		want    map[string][]string
	}{
		{
			name:    "nil",
			filters: Ec2Tags(nil),
			want:    map[string][]string{},
		},
		{
			name:    "nil",
			filters: Ec2Tags(map[string]string{}),
			want:    map[string][]string{},
		},
		{
			name: "tags",
			filters: Ec2Tags(map[string]string{
				"Name":    acctest.ResourcePrefix,
				"Purpose": "testing",
			}),
			want: map[string][]string{
				"tag:Name":    {acctest.ResourcePrefix},
				"tag:Purpose": {"testing"},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			got := testCase.filters.Map()

			testNameValuesFiltersVerifyMap(t, got, testCase.want)
		})
	}
}
