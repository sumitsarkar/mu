package common

import (
	"io"
	"time"
)

// Context defines the context object passed around
type Context struct {
	Config          Config
	StackManager    StackManager
	ClusterManager  ClusterManager
	InstanceManager InstanceManager
	ElbManager      ElbManager
	RdsManager      RdsManager
	ParamManager    ParamManager
	PipelineManager PipelineManager
	LogsManager     LogsManager
	DockerManager   DockerManager
	DockerOut       io.Writer
	TaskManager     TaskManager
	ArtifactManager ArtifactManager
}

// Config defines the structure of the yml file for the mu config
type Config struct {
	Environments []Environment `yaml:"environments,omitempty"`
	Service      Service       `yaml:"service,omitempty"`
	Basedir      string        `yaml:"-"`
	RelMuFile    string        `yaml:"-"`
	Repo         struct {
		Name     string
		Slug     string
		Revision string
		Branch   string
		Provider string
	} `yaml:"-"`
	Templates map[string]interface{} `yaml:"templates,omitempty"`
}

// Environment defines the structure of the yml file for an environment
type Environment struct {
	Name         string      `yaml:"name,omitempty"`
	Provider     EnvProvider `yaml:"provider,omitempty"`
	Loadbalancer struct {
		HostedZone  string `yaml:"hostedzone,omitempty"`
		Name        string `yaml:"name,omitempty"`
		Certificate string `yaml:"certificate,omitempty"`
		Internal    bool   `yaml:"internal,omitempty"`
	} `yaml:"loadbalancer,omitempty"`
	Cluster struct {
		InstanceType      string `yaml:"instanceType,omitempty"`
		ImageID           string `yaml:"imageId,omitempty"`
		InstanceTenancy   string `yaml:"instanceTenancy,omitempty"`
		DesiredCapacity   int    `yaml:"desiredCapacity,omitempty"`
		MaxSize           int    `yaml:"maxSize,omitempty"`
		KeyName           string `yaml:"keyName,omitempty"`
		SSHAllow          string `yaml:"sshAllow,omitempty"`
		ScaleOutThreshold int    `yaml:"scaleOutThreshold,omitempty"`
		ScaleInThreshold  int    `yaml:"scaleInThreshold,omitempty"`
		HTTPProxy         string `yaml:"httpProxy,omitempty"`
	} `yaml:"cluster,omitempty"`
	Discovery struct {
		Provider      string            `yaml:"provider,omitempty"`
		Configuration map[string]string `yaml:"configuration,omitempty"`
	} `yaml:"discovery,omitempty"`
	VpcTarget struct {
		VpcID             string   `yaml:"vpcId,omitempty"`
		InstanceSubnetIds []string `yaml:"instanceSubnetIds,omitempty"`
		ElbSubnetIds      []string `yaml:"elbSubnetIds,omitempty"`
	} `yaml:"vpcTarget,omitempty"`
}

type ScheduledTask struct {
	Expression 		string				   `yaml:"expression,omitempty"`
	Command			string				   `yaml:"command,omitempty"`
}

// Service defines the structure of the yml file for a service
type Service struct {
	Name            string                 `yaml:"name,omitempty"`
	DesiredCount    int                    `yaml:"desiredCount,omitempty"`
	Dockerfile      string                 `yaml:"dockerfile,omitempty"`
	ImageRepository string                 `yaml:"imageRepository,omitempty"`
	Port            int                    `yaml:"port,omitempty"`
	Protocol        string                 `yaml:"protocol,omitempty"`
	HealthEndpoint  string                 `yaml:"healthEndpoint,omitempty"`
	CPU             int                    `yaml:"cpu,omitempty"`
	Memory          int                    `yaml:"memory,omitempty"`
	Environment     map[string]interface{} `yaml:"environment,omitempty"`
	PathPatterns    []string               `yaml:"pathPatterns,omitempty"`
	HostPatterns    []string               `yaml:"hostPatterns,omitempty"`
	Priority        int                    `yaml:"priority,omitempty"`
	Pipeline        Pipeline               `yaml:"pipeline,omitempty"`
	Database        Database               `yaml:"database,omitempty"`
	ScheduledTasks   []ScheduledTask  	   `yaml:"scheduledTasks,omitempty"`
}

// Database definition
type Database struct {
	Name              string `yaml:"name,omitempty"`
	InstanceClass     string `yaml:"instanceClass,omitempty"`
	Engine            string `yaml:"engine,omitempty"`
	IamAuthentication bool   `yaml:"iamAuthentication,omitempty"`
	MasterUsername    string `yaml:"masterUsername,omitempty"`
	AllocatedStorage  string `yaml:"allocatedStorage,omitempty"`
}

// Pipeline definition
type Pipeline struct {
	Source struct {
		Provider string `yaml:"provider,omitempty"`
		Repo     string `yaml:"repo,omitempty"`
		Branch   string `yaml:"branch,omitempty"`
	} `yaml:"source,omitempty"`
	Build struct {
		Disabled    bool   `yaml:"disabled,omitempty"`
		Type        string `yaml:"type,omitempty"`
		ComputeType string `yaml:"computeType,omitempty"`
		Image       string `yaml:"image,omitempty"`
	} `yaml:"build,omitempty"`
	Acceptance struct {
		Disabled    bool   `yaml:"disabled,omitempty"`
		Environment string `yaml:"environment,omitempty"`
		Type        string `yaml:"type,omitempty"`
		ComputeType string `yaml:"computeType,omitempty"`
		Image       string `yaml:"image,omitempty"`
	} `yaml:"acceptance,omitempty"`
	Production struct {
		Disabled    bool   `yaml:"disabled,omitempty"`
		Environment string `yaml:"environment,omitempty"`
	} `yaml:"production,omitempty"`
	MuBaseurl string `yaml:"muBaseurl,omitempty"`
	MuVersion string `yaml:"muVersion,omitempty"`
}

// Stack summary
type Stack struct {
	ID             string
	Name           string
	Status         string
	StatusReason   string
	LastUpdateTime time.Time
	Tags           map[string]string
	Outputs        map[string]string
	Parameters     map[string]string
}

const (
	// StackStatusCreateInProgress is a StackStatus enum value
	StackStatusCreateInProgress = "CREATE_IN_PROGRESS"

	// StackStatusCreateFailed is a StackStatus enum value
	StackStatusCreateFailed = "CREATE_FAILED"

	// StackStatusCreateComplete is a StackStatus enum value
	StackStatusCreateComplete = "CREATE_COMPLETE"

	// StackStatusRollbackInProgress is a StackStatus enum value
	StackStatusRollbackInProgress = "ROLLBACK_IN_PROGRESS"

	// StackStatusRollbackFailed is a StackStatus enum value
	StackStatusRollbackFailed = "ROLLBACK_FAILED"

	// StackStatusRollbackComplete is a StackStatus enum value
	StackStatusRollbackComplete = "ROLLBACK_COMPLETE"

	// StackStatusDeleteInProgress is a StackStatus enum value
	StackStatusDeleteInProgress = "DELETE_IN_PROGRESS"

	// StackStatusDeleteFailed is a StackStatus enum value
	StackStatusDeleteFailed = "DELETE_FAILED"

	// StackStatusDeleteComplete is a StackStatus enum value
	StackStatusDeleteComplete = "DELETE_COMPLETE"

	// StackStatusUpdateInProgress is a StackStatus enum value
	StackStatusUpdateInProgress = "UPDATE_IN_PROGRESS"

	// StackStatusUpdateCompleteCleanupInProgress is a StackStatus enum value
	StackStatusUpdateCompleteCleanupInProgress = "UPDATE_COMPLETE_CLEANUP_IN_PROGRESS"

	// StackStatusUpdateComplete is a StackStatus enum value
	StackStatusUpdateComplete = "UPDATE_COMPLETE"

	// StackStatusUpdateRollbackInProgress is a StackStatus enum value
	StackStatusUpdateRollbackInProgress = "UPDATE_ROLLBACK_IN_PROGRESS"

	// StackStatusUpdateRollbackFailed is a StackStatus enum value
	StackStatusUpdateRollbackFailed = "UPDATE_ROLLBACK_FAILED"

	// StackStatusUpdateRollbackCompleteCleanupInProgress is a StackStatus enum value
	StackStatusUpdateRollbackCompleteCleanupInProgress = "UPDATE_ROLLBACK_COMPLETE_CLEANUP_IN_PROGRESS"

	// StackStatusUpdateRollbackComplete is a StackStatus enum value
	StackStatusUpdateRollbackComplete = "UPDATE_ROLLBACK_COMPLETE"

	// StackStatusReviewInProgress is a StackStatus enum value
	StackStatusReviewInProgress = "REVIEW_IN_PROGRESS"
)

// StackType describes supported stack types
type StackType string

// List of valid stack types
const (
	StackTypeVpc          StackType = "vpc"
	StackTypeTarget                 = "target"
	StackTypeEnv                    = "environment"
	StackTypeLoadBalancer           = "loadbalancer"
	StackTypeConsul                 = "consul"
	StackTypeRepo                   = "repo"
	StackTypeApp                    = "app"
	StackTypeService                = "service"
	StackTypePipeline               = "pipeline"
	StackTypeDatabase               = "database"
	StackTypeBucket                 = "bucket"
)

// EnvProvider describes supported environment strategies
type EnvProvider string

// List of valid environment strategies
const (
	EnvProviderEcs EnvProvider = "ecs"
	EnvProviderEc2             = "ec2"
)

// ArtifactProvider describes supported artifact strategies
type ArtifactProvider string

// List of valid artifact providers
const (
	ArtifactProviderEcr ArtifactProvider = "ecr"
	ArtifactProviderS3                   = "s3"
)

// Container describes container details
type Container struct {
	Name     string
	Instance string
}

// Task describes task definition
type Task struct {
	Name           string
	Environment    string
	Service        string
	TaskDefinition string
	Cluster        string
	Command        []string
	Containers     []Container
}

// JSONOutput common json definition
type JSONOutput struct {
	Values [1]struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	} `json:"values"`
}

// Int64Value returns the value of the int64 pointer passed in or
// 0 if the pointer is nil.
func Int64Value(v *int64) int64 {
	if v != nil {
		return *v
	}
	return 0
}

// StringValue returns the value of the string pointer passed in or
// "" if the pointer is nil.
func StringValue(v *string) string {
	if v != nil {
		return *v
	}
	return ""
}

// BoolValue returns the value of the bool pointer passed in or
// false if the pointer is nil.
func BoolValue(v *bool) bool {
	if v != nil {
		return *v
	}
	return false
}

// TimeValue returns the value of the time.Time pointer passed in or
// time.Time{} if the pointer is nil.
func TimeValue(v *time.Time) time.Time {
	if v != nil {
		return *v
	}
	return time.Time{}
}
