/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package awstasks

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"k8s.io/klog/v2"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraform"
	"k8s.io/kops/upup/pkg/fi/cloudup/terraformWriter"
)

// +kops:fitask
type InternetGateway struct {
	Name      *string
	Lifecycle fi.Lifecycle

	ID  *string
	VPC *VPC
	// Shared is set if this is a shared InternetGateway
	Shared *bool

	// Tags is a map of aws tags that are added to the InternetGateway
	Tags map[string]string
}

var _ fi.CompareWithID = &InternetGateway{}

func (e *InternetGateway) CompareWithID() *string {
	return e.ID
}

func findInternetGateway(ctx context.Context, cloud awsup.AWSCloud, request *ec2.DescribeInternetGatewaysInput) (*ec2types.InternetGateway, error) {
	response, err := cloud.EC2().DescribeInternetGateways(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("error listing InternetGateways: %v", err)
	}
	if response == nil || len(response.InternetGateways) == 0 {
		return nil, nil
	}

	if len(response.InternetGateways) != 1 {
		return nil, fmt.Errorf("found multiple InternetGateways matching tags")
	}
	igw := response.InternetGateways[0]
	return &igw, nil
}

func (e *InternetGateway) Find(c *fi.CloudupContext) (*InternetGateway, error) {
	ctx := c.Context()
	cloud := awsup.GetCloud(c)

	request := &ec2.DescribeInternetGatewaysInput{}

	shared := fi.ValueOf(e.Shared)
	if shared {
		if fi.ValueOf(e.VPC.ID) == "" {
			return nil, fmt.Errorf("VPC ID is required when InternetGateway is shared")
		}

		request.Filters = []ec2types.Filter{awsup.NewEC2Filter("attachment.vpc-id", *e.VPC.ID)}
	} else {
		if e.ID != nil {
			request.InternetGatewayIds = []string{fi.ValueOf(e.ID)}
		} else {
			request.Filters = cloud.BuildFilters(e.Name)
		}
	}

	igw, err := findInternetGateway(ctx, cloud, request)
	if err != nil {
		return nil, err
	}
	if igw == nil {
		return nil, nil
	}
	actual := &InternetGateway{
		ID:   igw.InternetGatewayId,
		Name: findNameTag(igw.Tags),
		Tags: intersectTags(igw.Tags, e.Tags),
	}

	klog.V(2).Infof("found matching InternetGateway %q", *actual.ID)

	for _, attachment := range igw.Attachments {
		actual.VPC = &VPC{ID: attachment.VpcId}
	}

	// Prevent spurious comparison failures
	actual.Shared = e.Shared
	actual.Lifecycle = e.Lifecycle
	if shared {
		actual.Name = e.Name
	}
	if e.ID == nil {
		e.ID = actual.ID
	}

	// We don't set the tags for a shared IGW
	if fi.ValueOf(e.Shared) {
		actual.Tags = e.Tags
	}

	return actual, nil
}

func (e *InternetGateway) Run(c *fi.CloudupContext) error {
	return fi.CloudupDefaultDeltaRunMethod(e, c)
}

func (s *InternetGateway) CheckChanges(a, e, changes *InternetGateway) error {
	if a != nil {
		// TODO: I think we can change it; we just detach & attach
		if changes.VPC != nil {
			return fi.CannotChangeField("VPC")
		}
	}

	return nil
}

func (_ *InternetGateway) RenderAWS(t *awsup.AWSAPITarget, a, e, changes *InternetGateway) error {
	ctx := context.TODO()
	shared := fi.ValueOf(e.Shared)
	if shared {
		// Verify the InternetGateway was found and matches our required settings
		if a == nil {
			return fmt.Errorf("InternetGateway for shared VPC was not found")
		}

		return nil
	}

	if a == nil {
		klog.V(2).Infof("Creating InternetGateway")

		request := &ec2.CreateInternetGatewayInput{
			TagSpecifications: awsup.EC2TagSpecification(ec2types.ResourceTypeInternetGateway, e.Tags),
		}

		response, err := t.Cloud.EC2().CreateInternetGateway(ctx, request)
		if err != nil {
			return fmt.Errorf("error creating InternetGateway: %v", err)
		}

		e.ID = response.InternetGateway.InternetGatewayId
	}

	if a == nil || (changes != nil && changes.VPC != nil) {
		klog.V(2).Infof("Creating InternetGatewayAttachment")

		attachRequest := &ec2.AttachInternetGatewayInput{
			VpcId:             e.VPC.ID,
			InternetGatewayId: e.ID,
		}

		_, err := t.Cloud.EC2().AttachInternetGateway(ctx, attachRequest)
		if err != nil {
			return fmt.Errorf("error attaching InternetGateway to VPC: %v", err)
		}
	}

	return t.AddAWSTags(*e.ID, e.Tags)
}

type terraformInternetGateway struct {
	VPCID *terraformWriter.Literal `cty:"vpc_id"`
	Tags  map[string]string        `cty:"tags"`
}

func (_ *InternetGateway) RenderTerraform(t *terraform.TerraformTarget, a, e, changes *InternetGateway) error {
	ctx := context.TODO()
	shared := fi.ValueOf(e.Shared)
	if shared {
		// Not terraform owned / managed

		// But ... attempt to discover the ID so TerraformLink works
		if e.ID == nil {
			request := &ec2.DescribeInternetGatewaysInput{}
			vpcID := fi.ValueOf(e.VPC.ID)
			if vpcID == "" {
				return fmt.Errorf("VPC ID is required when InternetGateway is shared")
			}
			request.Filters = []ec2types.Filter{awsup.NewEC2Filter("attachment.vpc-id", vpcID)}
			igw, err := findInternetGateway(ctx, t.Cloud.(awsup.AWSCloud), request)
			if err != nil {
				return err
			}
			if igw == nil {
				klog.Warningf("Cannot find internet gateway for VPC %q", vpcID)
			} else {
				e.ID = igw.InternetGatewayId
			}
		}

		return nil
	}

	tf := &terraformInternetGateway{
		VPCID: e.VPC.TerraformLink(),
		Tags:  e.Tags,
	}

	return t.RenderResource("aws_internet_gateway", *e.Name, tf)
}

func (e *InternetGateway) TerraformLink() *terraformWriter.Literal {
	shared := fi.ValueOf(e.Shared)
	if shared {
		if e.ID == nil {
			klog.Fatalf("ID must be set, if InternetGateway is shared: %s", e)
		}

		klog.V(4).Infof("reusing existing InternetGateway with id %q", *e.ID)
		return terraformWriter.LiteralFromStringValue(*e.ID)
	}

	return terraformWriter.LiteralProperty("aws_internet_gateway", *e.Name, "id")
}
