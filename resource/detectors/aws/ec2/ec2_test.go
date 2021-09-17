package ec2

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/logicmonitor/lm-telemetry-sdk-go/mock"
	"github.com/logicmonitor/lm-telemetry-sdk-go/utils"
	"go.opentelemetry.io/otel/sdk/resource"
)

var (
	localIP   = "127.0.0.1"
	nameTag   = "test-name-tag"
	testError = errors.New("test error")
)

func CreateEC2InstanceIdentityDocumentMock(PrivateIP string, err error) func() (ec2metadata.EC2InstanceIdentityDocument, error) {
	var getEc2InstanceIdentityDocumentSuccesMock = func() (ec2metadata.EC2InstanceIdentityDocument, error) {
		return ec2metadata.EC2InstanceIdentityDocument{
			PrivateIP: PrivateIP,
		}, err
	}
	return getEc2InstanceIdentityDocumentSuccesMock
}

func CreateGetNameTagMock(name string) func(instanceID, region string) string {
	var GetNameTagMock = func(instanceID, region string) string {
		return name
	}
	return GetNameTagMock
}

func CreateAWSSessionMock(sess *session.Session, err error) func(cfgs ...*aws.Config) (*session.Session, error) {
	return func(cfgs ...*aws.Config) (*session.Session, error) {
		return sess, err
	}
}

func TestDetect(t *testing.T) {
	oldgetEc2InstanceIdentityDocument := getEc2InstanceIdentityDocument
	//oldGetNameTag := GetNameTag
	oldutilsGetServiceDetails := utils.GetServiceDetails
	oldAddEnvResAttributes := utils.AddEnvResAttributes
	defer func() {
		getEc2InstanceIdentityDocument = oldgetEc2InstanceIdentityDocument
		//GetNameTag = oldGetNameTag
		utils.GetServiceDetails = oldutilsGetServiceDetails
		utils.AddEnvResAttributes = oldAddEnvResAttributes
	}()
	getEc2InstanceIdentityDocument = CreateEC2InstanceIdentityDocumentMock(localIP, nil)
	//GetNameTag = CreateGetNameTagMock(nameTag)
	utils.GetServiceDetails = mock.CreateGetServiceDetailsMock(map[string]string{"attrib1": "value1"})
	utils.AddEnvResAttributes = mock.CreateAddEnvResAttributesMock(resource.Empty(), nil)

	ec2Mock := mock.DetectorMock{
		Res: resource.Empty(),
		Err: nil,
	}

	ec2 := EC2{
		otelAWS: ec2Mock,
	}

	t.Run("Success test Case", func(t *testing.T) {

		_, err := ec2.Detect(context.Background())
		if err != ec2Mock.Err {
			t.Fatalf("Failed as err is not matching")
		}
	})

	t.Run("Error in adding environment resource attribute", func(t *testing.T) {
		oldAddEnvResAttributes := utils.AddEnvResAttributes
		defer func() {
			utils.AddEnvResAttributes = oldAddEnvResAttributes
		}()
		utils.AddEnvResAttributes = mock.CreateAddEnvResAttributesMock(resource.Empty(), testError)

		_, err := ec2.Detect(context.Background())
		if err != ec2Mock.Err {
			t.Fatalf("Failed as err is not matching")
		}
	})

	t.Run("Error in resource detection", func(t *testing.T) {
		ec2Mock := mock.DetectorMock{
			Res: resource.Empty(),
			Err: testError,
		}

		ec2 := EC2{
			otelAWS: ec2Mock,
		}

		_, err := ec2.Detect(context.Background())
		if err != ec2Mock.Err {
			t.Fatalf("Failed as err is not matching")
		}
	})

	t.Run("Resource is nil", func(t *testing.T) {
		ec2Mock := mock.DetectorMock{
			Res: nil,
			Err: nil,
		}

		ec2 := EC2{
			otelAWS: ec2Mock,
		}

		_, err := ec2.Detect(context.Background())
		if err != ec2Mock.Err {
			t.Fatalf("Failed as err is not matching")
		}
	})
}
