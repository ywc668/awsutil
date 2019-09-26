package awsutil

import (
	"runtime"
	"strings"

	log "github.com/sirupsen/logrus"

	//"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/ec2"

	jmespath "github.com/jmespath/go-jmespath"
)

func GetFuncName() string {
	pc := make([]uintptr, 1)  // at least 1 entry needed
	runtime.Callers(2, pc)
	/*https://github.com/golang/go/issues/19426
	f := runtime.FuncForPC(pc[0])
	n := f.Name()*/
	frm, _ := runtime.CallersFrames(pc).Next()
	n := frm.Function
	return n[strings.LastIndex(n, ".")+1:]
}

func ecsErrorHandler(err error) {
	if aerr, ok := err.(awserr.Error); ok {
		switch aerr.Code() {
		case ecs.ErrCodeServerException:
			log.Error(ecs.ErrCodeServerException, aerr.Error())
		case ecs.ErrCodeClientException:
			log.Error(ecs.ErrCodeClientException, aerr.Error())
		case ecs.ErrCodeInvalidParameterException:
			log.Error(ecs.ErrCodeInvalidParameterException, aerr.Error())
		case ecs.ErrCodeClusterNotFoundException:
			log.Error(ecs.ErrCodeClusterNotFoundException, aerr.Error())
		default:
			log.Error(aerr.Error())
		}
	} else {
		// Print the error, cast err to awserr.Error to get the Code and
		// Message from an error.
		log.Error(err.Error())
	}
}

func jmespathSearch(query string, data interface{}) (r interface{}) {
	if data == nil {
		return
	}
	r, e := jmespath.Search(query, data)
	if e != nil {
		log.Error(e)
		r = nil
	}
	return
}

func awsAPIWrapper(wrapper func() (interface{}, error)) (r interface{}) {
	r, e := wrapper()
	if e != nil {
		ecsErrorHandler(e)
		r = nil
	}
	return
}

// New style of aws api below, different from above, using jmespath

func (a AWSUtil) ECSListClusters(query string) interface{} {
	output := awsAPIWrapper(func() (interface{}, error) {
		input := &ecs.ListClustersInput{}
		return a.ecssvc.ListClusters(input)
	})

	return jmespathSearch(query, output)
}

func (a AWSUtil) ECSListServices(cluster *string, query string) interface{} {
	output := awsAPIWrapper(func() (interface{}, error) {
		input := &ecs.ListServicesInput{Cluster: cluster}
		return a.ecssvc.ListServices(input)
	})

	return jmespathSearch(query, output)
}

func (a AWSUtil) ECSDescribeServices(cluster *string, serviceIDs []*string, query string) interface{} {
	output := awsAPIWrapper(func() (interface{}, error) {
		input := &ecs.DescribeServicesInput{
			Cluster: cluster,
			Services: serviceIDs,
		}
		return a.ecssvc.DescribeServices(input)
	})

	return jmespathSearch(query, output)
}

func (a AWSUtil) ECSDescribeTaskDefinition(taskDefinition *string, query string) interface{} {
	output := awsAPIWrapper(func() (interface{}, error) {
		input := &ecs.DescribeTaskDefinitionInput{
			TaskDefinition: taskDefinition,
		}
		return a.ecssvc.DescribeTaskDefinition(input)
	})

	return jmespathSearch(query, output)
}

func (a AWSUtil) ECSListContainerInstances(cluster, filter *string, query string) interface{} {
	output := awsAPIWrapper(func() (interface{}, error) {
		input := &ecs.ListContainerInstancesInput{
			Cluster: cluster,
			Filter: filter,
		}
		return a.ecssvc.ListContainerInstances(input)
	})

	return jmespathSearch(query, output)
}

func (a AWSUtil) EC2DescribeInstances(ec2IDs []*string, query string) interface{} {
	output := awsAPIWrapper(func() (interface{}, error) {
		input := &ec2.DescribeInstancesInput{
			InstanceIds: ec2IDs,
		}
		return a.ec2svc.DescribeInstances(input)
	})

	return jmespathSearch(query, output)
}

func (a AWSUtil) ECSDescribeContainerInstances(cluster *string, containerInstanceIDs []*string, query string) interface{} {
	output := awsAPIWrapper(func() (interface{}, error) {
		input := &ecs.DescribeContainerInstancesInput{
			Cluster: cluster,
			ContainerInstances: containerInstanceIDs,
		}
		return a.ecssvc.DescribeContainerInstances(input)
	})

	return jmespathSearch(query, output)
}

func (a AWSUtil) ECSListTasks(cluster *string, service *string, query string) interface{} {
	output := awsAPIWrapper(func() (interface{}, error) {
		input := &ecs.ListTasksInput{
			Cluster: cluster,
			ServiceName: service,
		}
		return a.ecssvc.ListTasks(input)
	})

	return jmespathSearch(query, output)
}

func (a AWSUtil) ECSDescribeTasks(cluster *string, tasks []*string, query string) interface{} {
	output := awsAPIWrapper(func() (interface{}, error) {
		input := &ecs.DescribeTasksInput{
			Cluster: cluster,
			Tasks: tasks,
		}
		return a.ecssvc.DescribeTasks(input)
	})

	return jmespathSearch(query, output)
}
