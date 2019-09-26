package awsutil

import (
	"fmt"
	"strings"
	"encoding/json"
	"net/http"
	"io/ioutil"

	log "github.com/sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/ec2"
)

type AWSUtil struct {
	ecssvc		*ecs.ECS
	ec2svc		*ec2.EC2
}

func New(region string) AWSUtil {
	a := AWSUtil{}
	if region == "" {
		region = getRegionFromMetaData()
	}

	// If no region found from EC2 metadata, fallback to obtain region by
	// traditional ways, such as AWS_REGION, config/credential file, etc.
	awscfg := []*aws.Config(nil)
	if region != "" {
		awscfg = make([]*aws.Config, 1, 1)
		awscfg[0] = &aws.Config{Region: aws.String(region)}
		//awscfg = append(awscfg, &aws.Config{Region: aws.String(region))
	}
	session, err := session.NewSession(awscfg...)
	if err != nil {
		err = fmt.Errorf("Failed to create AWS session: %s", err)
		log.Fatal(err)
	}

	a.ecssvc = ecs.New(session)
	a.ec2svc = ec2.New(session)
	return a
}

func getRegionFromMetaData() (region string) {
	url := "http://169.254.169.254/latest/dynamic/instance-identity/document/"
	resp, err := http.Get(url)
	if err != nil {
		log.Warnf("Failed to get region through URL %s, ", url, err)
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		log.Warnf("Failed to get region due to failing to read http body, %s", err)
		return
	}

	data := struct{ Region string }{}
	if err := json.Unmarshal(body, &data); err != nil {
		log.Warnf("Failed to get region because JSON unmarshaling failed: %s", err)
		return
	}
	log.Debug(data.Region)
	return data.Region
}

func sliceInterfaceToStringPointer(ifaces []interface{}) []*string {
	result := make([]*string, len(ifaces))
	for i, s := range ifaces {
		var ok bool
		result[i], ok = s.(*string)
		if !ok {
			str := s.(string)
			result[i] = &str
		}
	}
	return result
}

func (a AWSUtil) GetContainerInstanceIDs(cluster, service string) []*string {
	filter := "task:group == service:" + service
	return a.ECSListContainerInstances(&cluster, &filter, "ContainerInstanceArns").([]*string)
}

func (a AWSUtil) GetContainerInstanceEC2IDs(cluster string, containerInstanceIDs []*string) []*string {
	if containerInstanceIDs == nil {
		return nil
	}
	return sliceInterfaceToStringPointer(
		a.ECSDescribeContainerInstances(&cluster, containerInstanceIDs, "ContainerInstances[*].Ec2InstanceId").([]interface{}))
}

func (a AWSUtil) GetEC2InstancePrivateIPs(ec2InstanceIDs []*string) []*string {
	if ec2InstanceIDs == nil {
		return nil
	}
	return sliceInterfaceToStringPointer(
		a.EC2DescribeInstances(ec2InstanceIDs, "Reservations[*].Instances[0].PrivateIpAddress").([]interface{}))
}

func (a AWSUtil) GetECSClusters() []string {
	clusters := a.ECSListClusters("ClusterArns").([]*string)
	result := make([]string, len(clusters))
	for i, c := range clusters {
		result[i] = (*c)[strings.LastIndex(*c, "/")+1:]
	}
	log.Debug(result)
	return result
}
