// Code generated by smithy-go-codegen DO NOT EDIT.

package ssm

import (
	"context"
	"fmt"
	awsmiddleware "github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/aws/smithy-go/middleware"
	smithyhttp "github.com/aws/smithy-go/transport/http"
)

// Get information about one or more parameters by specifying multiple parameter
// names. To get information about a single parameter, you can use the GetParameter
// operation instead.
func (c *Client) GetParameters(ctx context.Context, params *GetParametersInput, optFns ...func(*Options)) (*GetParametersOutput, error) {
	if params == nil {
		params = &GetParametersInput{}
	}

	result, metadata, err := c.invokeOperation(ctx, "GetParameters", params, optFns, c.addOperationGetParametersMiddlewares)
	if err != nil {
		return nil, err
	}

	out := result.(*GetParametersOutput)
	out.ResultMetadata = metadata
	return out, nil
}

type GetParametersInput struct {

	// The names or Amazon Resource Names (ARNs) of the parameters that you want to
	// query. For parameters shared with you from another account, you must use the
	// full ARNs. To query by parameter label, use "Name": "name:label" . To query by
	// parameter version, use "Name": "name:version" . For more information about
	// shared parameters, see Working with shared parameters (https://docs.aws.amazon.com/systems-manager/latest/userguide/parameter-store-shared-parameters.html)
	// in the Amazon Web Services Systems Manager User Guide.
	//
	// This member is required.
	Names []string

	// Return decrypted secure string value. Return decrypted values for secure string
	// parameters. This flag is ignored for String and StringList parameter types.
	WithDecryption *bool

	noSmithyDocumentSerde
}

type GetParametersOutput struct {

	// A list of parameters that aren't formatted correctly or don't run during an
	// execution.
	InvalidParameters []string

	// A list of details for a parameter.
	Parameters []types.Parameter

	// Metadata pertaining to the operation's result.
	ResultMetadata middleware.Metadata

	noSmithyDocumentSerde
}

func (c *Client) addOperationGetParametersMiddlewares(stack *middleware.Stack, options Options) (err error) {
	if err := stack.Serialize.Add(&setOperationInputMiddleware{}, middleware.After); err != nil {
		return err
	}
	err = stack.Serialize.Add(&awsAwsjson11_serializeOpGetParameters{}, middleware.After)
	if err != nil {
		return err
	}
	err = stack.Deserialize.Add(&awsAwsjson11_deserializeOpGetParameters{}, middleware.After)
	if err != nil {
		return err
	}
	if err := addProtocolFinalizerMiddlewares(stack, options, "GetParameters"); err != nil {
		return fmt.Errorf("add protocol finalizers: %v", err)
	}

	if err = addlegacyEndpointContextSetter(stack, options); err != nil {
		return err
	}
	if err = addSetLoggerMiddleware(stack, options); err != nil {
		return err
	}
	if err = addClientRequestID(stack); err != nil {
		return err
	}
	if err = addComputeContentLength(stack); err != nil {
		return err
	}
	if err = addResolveEndpointMiddleware(stack, options); err != nil {
		return err
	}
	if err = addComputePayloadSHA256(stack); err != nil {
		return err
	}
	if err = addRetry(stack, options); err != nil {
		return err
	}
	if err = addRawResponseToMetadata(stack); err != nil {
		return err
	}
	if err = addRecordResponseTiming(stack); err != nil {
		return err
	}
	if err = addClientUserAgent(stack, options); err != nil {
		return err
	}
	if err = smithyhttp.AddErrorCloseResponseBodyMiddleware(stack); err != nil {
		return err
	}
	if err = smithyhttp.AddCloseResponseBodyMiddleware(stack); err != nil {
		return err
	}
	if err = addSetLegacyContextSigningOptionsMiddleware(stack); err != nil {
		return err
	}
	if err = addOpGetParametersValidationMiddleware(stack); err != nil {
		return err
	}
	if err = stack.Initialize.Add(newServiceMetadataMiddleware_opGetParameters(options.Region), middleware.Before); err != nil {
		return err
	}
	if err = addRecursionDetection(stack); err != nil {
		return err
	}
	if err = addRequestIDRetrieverMiddleware(stack); err != nil {
		return err
	}
	if err = addResponseErrorMiddleware(stack); err != nil {
		return err
	}
	if err = addRequestResponseLogging(stack, options); err != nil {
		return err
	}
	if err = addDisableHTTPSMiddleware(stack, options); err != nil {
		return err
	}
	return nil
}

func newServiceMetadataMiddleware_opGetParameters(region string) *awsmiddleware.RegisterServiceMetadata {
	return &awsmiddleware.RegisterServiceMetadata{
		Region:        region,
		ServiceID:     ServiceID,
		OperationName: "GetParameters",
	}
}