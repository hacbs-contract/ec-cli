// Copyright 2022 Red Hat, Inc.
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
//
// SPDX-License-Identifier: Apache-2.0

package policy

import (
	"context"
	"errors"
	"strings"

	"github.com/open-policy-agent/conftest/output"
	"github.com/open-policy-agent/conftest/parser"
	"github.com/open-policy-agent/conftest/policy"
	"github.com/sigstore/cosign/pkg/oci"
	"k8s.io/apimachinery/pkg/types"
)

type policyEvaluator struct {
	policyName types.NamespacedName
	k8s        kubernetes
	source     policySource
	out        output.Outputter
}

var kubernetesCreator = NewKubernetes

// NewPolicyEvaluator constructs a policyEvaluator that evaluates according to the pointed at policyConfiguration
func NewPolicyEvaluator(policyConfiguration string) (*policyEvaluator, error) {
	if policyConfiguration == "" {
		return nil, errors.New("policy: policy name is required")
	}

	policyName := types.NamespacedName{
		Name: policyConfiguration,
	}
	policyParts := strings.SplitN(policyConfiguration, string(types.Separator), 2)
	if len(policyParts) == 2 {
		policyName = types.NamespacedName{
			Namespace: policyParts[0],
			Name:      policyParts[1],
		}
	}

	k, err := kubernetesCreator()
	if err != nil {
		return nil, err
	}

	return &policyEvaluator{
		policyName: policyName,
		k8s:        *k,
		source:     NewPolicySource(),
		out: output.Get("json", output.Options{
			NoColor:            true,
			SuppressExceptions: false,
			Tracing:            false,
			JUnitHideMessage:   true,
		}),
	}, nil
}

func (p *policyEvaluator) Evaluate(ctx context.Context, attestations []oci.Signature) ([]output.CheckResult, error) {
	ecp, err := p.k8s.fetchEnterpriseContractPolicy(ctx, p.policyName)
	if err != nil {
		return nil, err
	}

	policies, err := p.source.fetchPolicySources(ctx, ecp.Spec)
	if err != nil {
		return nil, err
	}

	data, err := fetchPolicyData(ecp.Spec)
	if err != nil {
		return nil, err
	}

	inputs, err := fetchInputData(ctx, attestations)
	if err != nil {
		return nil, err
	}

	configurations, err := parser.ParseConfigurations(inputs)
	if err != nil {
		return nil, err
	}

	engine, err := policy.LoadWithData(ctx, policies, []string{data}, "")
	if err != nil {
		return nil, err
	}

	results, err := engine.Check(ctx, configurations, "main")
	if err != nil {
		return nil, err
	}

	return results, nil
}

func (p *policyEvaluator) Output(results []output.CheckResult) error {
	err := p.out.Output(results)

	return err
}
