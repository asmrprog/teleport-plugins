// Copyright 2023 Gravitational, Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package test

import (
	"time"

	"github.com/gravitational/teleport/api/types"
	"github.com/gravitational/trace"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/require"
)

func (s *TerraformSuite) TestOktaImportRule() {
	checkDestroyed := func(state *terraform.State) error {
		_, err := s.client.OktaClient().GetOktaImportRule(s.Context(), "test")
		if trace.IsNotFound(err) {
			return nil
		}

		return err
	}

	name := "teleport_okta_import_rule.test"

	resource.Test(s.T(), resource.TestCase{
		ProtoV6ProviderFactories: s.terraformProviders,
		CheckDestroy:             checkDestroyed,
		Steps: []resource.TestStep{
			{
				Config: s.getFixture("okta_import_rule_0_create.tf"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "kind", "okta_import_rule"),
				),
			},
			{
				Config:   s.getFixture("okta_import_rule_0_create.tf"),
				PlanOnly: true,
			},
			{
				Config: s.getFixture("okta_import_rule_1_update.tf"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "kind", "okta_import_rule"),
				),
			},
			{
				Config:   s.getFixture("okta_import_rule_1_update.tf"),
				PlanOnly: true,
			},
		},
	})
}

func (s *TerraformSuite) TestImportOktaImportRule() {
	r := "teleport_okta_import_rule"
	id := "test_import"
	name := r + "." + id

	oir := &types.OktaImportRuleV1{
		ResourceHeader: types.ResourceHeader{
			Metadata: types.Metadata{
				Name: id,
			},
		},
		Spec: types.OktaImportRuleSpecV1{
			Priority: 100,
			Mappings: []*types.OktaImportRuleMappingV1{
				{
					AddLabels: map[string]string{
						"label1": "value1",
					},
					Match: []*types.OktaImportRuleMatchV1{
						{
							AppIDs: []string{"1", "2", "3"},
						},
					},
				},
				{
					AddLabels: map[string]string{
						"label2": "value2",
					},
					Match: []*types.OktaImportRuleMatchV1{
						{
							GroupIDs: []string{"1", "2", "3"},
						},
					},
				},
			},
		},
	}
	err := oir.CheckAndSetDefaults()
	require.NoError(s.T(), err)

	_, err = s.client.OktaClient().CreateOktaImportRule(s.Context(), oir)
	require.NoError(s.T(), err)

	require.Eventually(s.T(), func() bool {
		_, err := s.client.OktaClient().GetOktaImportRule(s.Context(), oir.GetName())
		if trace.IsNotFound(err) {
			return false
		}
		require.NoError(s.T(), err)
		return true
	}, 5*time.Second, time.Second)

	resource.Test(s.T(), resource.TestCase{
		ProtoV6ProviderFactories: s.terraformProviders,
		Steps: []resource.TestStep{
			{
				Config:        s.terraformConfig + "\n" + `resource "` + r + `" "` + id + `" { }`,
				ResourceName:  name,
				ImportState:   true,
				ImportStateId: id,
				ImportStateCheck: func(state []*terraform.InstanceState) error {
					require.Equal(s.T(), state[0].Attributes["kind"], "okta_import_rule")

					return nil
				},
			},
		},
	})
}
