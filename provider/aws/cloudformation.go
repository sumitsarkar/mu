package aws

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/cloudformation/cloudformationiface"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
	"github.com/briandowns/spinner"
	"github.com/stelligent/mu/common"
	"golang.org/x/crypto/ssh/terminal"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"
)

type cloudformationStackManager struct {
	dryrun           bool
	skipVersionCheck bool
	cfnAPI           cloudformationiface.CloudFormationAPI
	ec2API           ec2iface.EC2API
}

// NewStackManager creates a new StackManager backed by cloudformation
func newStackManager(sess *session.Session, dryrun bool, skipVersionCheck bool) (common.StackManager, error) {
	log.Debug("Connecting to CloudFormation service")
	cfnAPI := cloudformation.New(sess)

	log.Debug("Connecting to EC2 service")
	ec2API := ec2.New(sess)

	return &cloudformationStackManager{
		dryrun:           dryrun,
		skipVersionCheck: skipVersionCheck,
		cfnAPI:           cfnAPI,
		ec2API:           ec2API,
	}, nil

}

func buildStackParameters(parameters map[string]string) []*cloudformation.Parameter {
	stackParameters := make([]*cloudformation.Parameter, 0, len(parameters))
	for key, value := range parameters {
		stackParameters = append(stackParameters,
			&cloudformation.Parameter{
				ParameterKey:     aws.String(key),
				ParameterValue:   aws.String(value),
				UsePreviousValue: aws.Bool(value == ""),
			})
	}
	return stackParameters
}

func buildStackTags(tags map[string]string) []*cloudformation.Tag {
	stackTags := make([]*cloudformation.Tag, 0, len(tags)+2)

	stackTags = append(stackTags,
		&cloudformation.Tag{
			Key:   aws.String("mu:version"),
			Value: aws.String(common.GetVersion()),
		})

	for key, value := range tags {
		if value != "" {
			stackTags = append(stackTags,
				&cloudformation.Tag{
					Key:   aws.String(key),
					Value: aws.String(value),
				})
		}
	}
	return stackTags
}

// UpsertStack will create/update the cloudformation stack
func (cfnMgr *cloudformationStackManager) UpsertStack(stackName string, templateBodyReader io.Reader, parameters map[string]string, tags map[string]string, roleArn string) error {
	stack := cfnMgr.AwaitFinalStatus(stackName)

	// delete stack if in rollback status
	if stack != nil && stack.Status == cloudformation.StackStatusRollbackComplete {
		log.Warningf("  Stack '%s' was in '%s' status, deleting...", stackName, stack.Status)
		err := cfnMgr.DeleteStack(stackName)
		if err != nil {
			return err
		}
		stack = cfnMgr.AwaitFinalStatus(stackName)
	}

	// check if stack is incompatible
	if !cfnMgr.skipVersionCheck && stack != nil {
		oldMajorVersion, e1 := strconv.Atoi(strings.Split(stack.Tags["version"], ".")[0])
		newMajorVersion, e2 := strconv.Atoi(strings.Split(common.GetVersion(), ".")[0])

		if e1 != nil {
			log.Warningf("Unable to parse major number for existing stack: %s", stack.Tags["version"])
			log.Warningf("Unable to parse major number for mu: %s", common.GetVersion())
		}

		log.Debugf("comparing stack versions old:%d new:%d", oldMajorVersion, newMajorVersion)

		if e1 == nil && e2 == nil {
			if oldMajorVersion < newMajorVersion {
				return fmt.Errorf("Unable to upsert stack '%s' with existing version '%s' to newer version '%s' (can be overriden with -F)", stackName, stack.Tags["version"], common.GetVersion())
			}
			if oldMajorVersion > newMajorVersion {
				return fmt.Errorf("Unable to upsert stack '%s' with existing version '%s' to older version '%s' (can be overridden with -F)", stackName, stack.Tags["version"], common.GetVersion())
			}
		}
	}

	// load the template
	templateBodyBytes := new(bytes.Buffer)
	templateBodyBytes.ReadFrom(templateBodyReader)
	templateBody := aws.String(templateBodyBytes.String())

	// stack parameters
	stackParameters := buildStackParameters(parameters)

	// stack tags
	stackTags := buildStackTags(tags)

	// directory to write cfn to
	cfnDirectory := fmt.Sprintf("%s/mu-cloudformation", os.TempDir())

	cfnAPI := cfnMgr.cfnAPI
	if stack == nil || stack.Status == "" {

		log.Debugf("  Creating stack named '%s'", stackName)
		log.Debugf("  Stack parameters:\n\t%s", stackParameters)
		log.Debugf("  Assume role:\n\t%s", roleArn)
		log.Debugf("  Stack tags:\n\t%s", stackTags)
		params := &cloudformation.CreateStackInput{
			StackName:    aws.String(stackName),
			Parameters:   stackParameters,
			TemplateBody: templateBody,
			Tags:         stackTags,
		}

		if roleArn != "" {
			params.RoleARN = aws.String(roleArn)
		}

		if tags["mu:type"] == string(common.StackTypeIam) {
			params.Capabilities = []*string{
				aws.String(cloudformation.CapabilityCapabilityNamedIam),
			}
		}

		if cfnMgr.dryrun {
			writeTemplateAndConfig(cfnDirectory, stackName, templateBodyBytes, parameters)
			log.Infof("  DRYRUN: Skipping create of stack named '%s'.  Template and parameters written to '%s'", stackName, cfnDirectory)
			return nil
		}

		_, err := cfnAPI.CreateStack(params)
		log.Debug("  Create stack complete err=%s", err)
		if err != nil {
			return err
		}

		waitParams := &cloudformation.DescribeStacksInput{
			StackName: aws.String(stackName),
		}
		log.Debug("  Waiting for stack to exist...")
		cfnAPI.WaitUntilStackExists(waitParams)
		log.Debug("  Stack exists.")

	} else {
		log.Debugf("  Updating stack named '%s'", stackName)
		log.Debugf("  Prior state: %s", stack.Status)
		log.Debugf("  Stack parameters:\n\t%s", stackParameters)
		log.Debugf("  Stack tags:\n\t%s", stackTags)
		params := &cloudformation.UpdateStackInput{
			StackName:    aws.String(stackName),
			Parameters:   stackParameters,
			TemplateBody: templateBody,
			Tags:         stackTags,
		}

		if roleArn != "" {
			params.RoleARN = aws.String(roleArn)
		}

		if tags["mu:type"] == string(common.StackTypeIam) {
			params.Capabilities = []*string{
				aws.String(cloudformation.CapabilityCapabilityNamedIam),
			}
		}

		if cfnMgr.dryrun {
			writeTemplateAndConfig(cfnDirectory, stackName, templateBodyBytes, parameters)
			log.Infof("  DRYRUN: Skipping update of stack named '%s'.  Template and parameters written to '%s'", stackName, cfnDirectory)
			return nil
		}

		_, err := cfnAPI.UpdateStack(params)
		log.Debug("  Update stack complete err=%s", err)
		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok {
				if awsErr.Code() == "ValidationError" && awsErr.Message() == "No updates are to be performed." {
					log.Infof("  No changes for stack '%s'", stackName)
					return nil
				}
			}
			return err
		}

	}
	return nil
}

// AwaitFinalStatus waits for the stack to arrive in a final status
//  returns: final status, or empty string if stack doesn't exist
func (cfnMgr *cloudformationStackManager) AwaitFinalStatus(stackName string) *common.Stack {
	cfnAPI := cfnMgr.cfnAPI
	params := &cloudformation.DescribeStacksInput{
		StackName: aws.String(stackName),
	}

	// initialize Spinner
	var statusSpinner *spinner.Spinner
	if terminal.IsTerminal(int(os.Stdout.Fd())) {
		statusSpinner = spinner.New(spinner.CharSets[9], 100*time.Millisecond)
	}

	if statusSpinner != nil {
		statusSpinner.Start()
		defer statusSpinner.Stop()
	}

	var priorEventTime *time.Time
	for {

		resp, err := cfnAPI.DescribeStacks(params)

		if statusSpinner != nil {
			statusSpinner.Stop()
		}
		if err != nil || resp == nil || len(resp.Stacks) != 1 {
			log.Debugf("  Stack doesn't exist ... stack=%s", stackName)
			return nil
		}

		if !strings.HasSuffix(aws.StringValue(resp.Stacks[0].StackStatus), "_IN_PROGRESS") {
			log.Debugf("  Returning final status for stack:%s ... status=%s", stackName, *resp.Stacks[0].StackStatus)
			return buildStack(resp.Stacks[0])
		}

		eventParams := &cloudformation.DescribeStackEventsInput{
			StackName: aws.String(stackName),
		}
		eventResp, err := cfnAPI.DescribeStackEvents(eventParams)
		if err == nil && eventResp != nil {
			numEvents := len(eventResp.StackEvents)
			for i := numEvents - 1; i >= 0; i-- {
				e := eventResp.StackEvents[i]
				if priorEventTime == nil || priorEventTime.Before(aws.TimeValue(e.Timestamp)) {

					status := aws.StringValue(e.ResourceStatus)
					eventMesg := fmt.Sprintf("  %s (%s) %s %s", aws.StringValue(e.LogicalResourceId),
						aws.StringValue(e.ResourceType),
						status,
						aws.StringValue(e.ResourceStatusReason))
					if strings.HasSuffix(status, "_IN_PROGRESS") {
						if statusSpinner != nil {
							statusSpinner.Suffix = eventMesg
							log.Debug(eventMesg)
						} else {
							log.Info(eventMesg)
						}
					} else if strings.HasSuffix(status, "_FAILED") {
						log.Error(eventMesg)
					} else {
						log.Debug(eventMesg)
					}
					priorEventTime = e.Timestamp
				}

			}

		}

		log.Debugf("  Not in final status (%s)...sleeping for 5 seconds", *resp.Stacks[0].StackStatus)
		if statusSpinner != nil {
			statusSpinner.Start()
		}
		time.Sleep(time.Second * 5)
	}
}

func buildStack(stackDetails *cloudformation.Stack) *common.Stack {
	stack := new(common.Stack)
	stack.ID = aws.StringValue(stackDetails.StackId)
	stack.Name = aws.StringValue(stackDetails.StackName)
	stack.Status = aws.StringValue(stackDetails.StackStatus)
	stack.StatusReason = aws.StringValue(stackDetails.StackStatusReason)
	if aws.TimeValue(stackDetails.LastUpdatedTime).Unix() > 0 {
		stack.LastUpdateTime = aws.TimeValue(stackDetails.LastUpdatedTime)
	} else {
		stack.LastUpdateTime = aws.TimeValue(stackDetails.CreationTime)
	}
	stack.Tags = make(map[string]string)
	stack.Outputs = make(map[string]string)
	stack.Parameters = make(map[string]string)

	for _, tag := range stackDetails.Tags {
		key := aws.StringValue(tag.Key)
		if strings.HasPrefix(key, "mu:") {
			stack.Tags[key[3:]] = aws.StringValue(tag.Value)
		}
	}

	for _, output := range stackDetails.Outputs {
		stack.Outputs[aws.StringValue(output.OutputKey)] = aws.StringValue(output.OutputValue)
	}

	for _, param := range stackDetails.Parameters {
		stack.Parameters[aws.StringValue(param.ParameterKey)] = aws.StringValue(param.ParameterValue)
	}

	return stack
}

// ListStacks will find mu stacks
func (cfnMgr *cloudformationStackManager) ListStacks(stackType common.StackType) ([]*common.Stack, error) {
	cfnAPI := cfnMgr.cfnAPI

	params := &cloudformation.DescribeStacksInput{}

	var stacks []*common.Stack

	log.Debugf("Searching for stacks of type '%s'", stackType)

	err := cfnAPI.DescribeStacksPages(params,
		func(page *cloudformation.DescribeStacksOutput, lastPage bool) bool {
			for _, stackDetails := range page.Stacks {
				if cloudformation.StackStatusDeleteComplete == aws.StringValue(stackDetails.StackStatus) {
					continue
				}

				stack := buildStack(stackDetails)

				if stack.Tags["type"] == string(stackType) {
					stacks = append(stacks, stack)
				}
			}

			return true
		})

	if err != nil {
		return nil, err
	}
	return stacks, nil
}

// GetStack get a specific stack
func (cfnMgr *cloudformationStackManager) GetStack(stackName string) (*common.Stack, error) {
	cfnAPI := cfnMgr.cfnAPI

	params := &cloudformation.DescribeStacksInput{StackName: aws.String(stackName)}

	log.Debugf("Searching for stack named '%s'", stackName)

	resp, err := cfnAPI.DescribeStacks(params)
	if err != nil {
		return nil, err
	}
	stack := buildStack(resp.Stacks[0])

	return stack, nil
}

// FindLatestImageID for a given
func (cfnMgr *cloudformationStackManager) FindLatestImageID(namePattern string) (string, error) {
	ec2Api := cfnMgr.ec2API
	resp, err := ec2Api.DescribeImages(&ec2.DescribeImagesInput{
		Owners: []*string{aws.String("amazon")},
		Filters: []*ec2.Filter{
			{
				Name: aws.String("name"),
				Values: []*string{
					aws.String(namePattern),
				},
			},
		},
	})

	if err != nil {
		return "", err
	}

	var imageID string
	var imageCreateDate time.Time
	for _, image := range resp.Images {
		createDate, err := time.Parse(time.RFC3339, aws.StringValue(image.CreationDate))
		if err != nil {
			return "", err
		}
		if imageCreateDate.Before(createDate) {
			imageCreateDate = createDate
			imageID = aws.StringValue(image.ImageId)
		}
	}

	if imageID == "" {
		return "", errors.New("Unable to find image")
	}
	log.Debugf("Found latest imageId %s for pattern %s", imageID, namePattern)
	return imageID, nil
}

// DeleteStack delete a specific stack
func (cfnMgr *cloudformationStackManager) DeleteStack(stackName string) error {
	cfnAPI := cfnMgr.cfnAPI

	params := &cloudformation.DeleteStackInput{StackName: aws.String(stackName)}

	if cfnMgr.dryrun {
		log.Infof("  DRYRUN: Skipping delete of stack named '%s'", stackName)
		return nil
	}

	log.Debugf("Deleting stack named '%s'", stackName)

	_, err := cfnAPI.DeleteStack(params)
	return err
}

func writeTemplateAndConfig(cfnDirectory string, stackName string, templateBodyBytes *bytes.Buffer, parameters map[string]string) error {
	os.MkdirAll(cfnDirectory, 0700)
	templateFile := fmt.Sprintf("%s/template-%s.yml", cfnDirectory, stackName)
	err := ioutil.WriteFile(templateFile, templateBodyBytes.Bytes(), 0600)
	if err != nil {
		return err
	}

	configMap := make(map[string]map[string]string)
	configMap["Parameters"] = parameters
	configBody, err := json.MarshalIndent(configMap, "", "  ")
	if err != nil {
		return err
	}
	configFile := fmt.Sprintf("%s/config-%s.json", cfnDirectory, stackName)
	err = ioutil.WriteFile(configFile, configBody, 0600)
	return nil
}
