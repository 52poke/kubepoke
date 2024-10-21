package s3policy

import (
	"context"
	"html/template"
	"log/slog"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/mudkipme/kubepoke/internal/interfaces"
)

const bucketPolicyTemplate = `{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Principal": "*",
            "Action": "s3:GetObject",
            "Resource": "arn:aws:s3:::{{.BucketName}}/*",
            "Condition": {
                "IpAddress": {
                    "aws:SourceIp": [
                        {{- range $index, $ip := .IPs }}
                        "{{ $ip }}"{{ if lt $index (sub (len $.IPs) 1) }},{{ end }}
                        {{- end }}
                    ]
                }
            }
        }
    ]
}`

type S3PolicyHelper struct {
	client *s3.Client
	bucket string
}

type S3PolicyHelperConfig struct {
	AccessKeyID     string `json:"accessKeyId"`
	SecretAccessKey string `json:"secretAccessKey"`
	Region          string `json:"region"`
	Endpoint        string `json:"endpoint"`
	BucketName      string `json:"bucketName"`
}

type configData struct {
	BucketName string
	IPs        []string
}

func NewS3PolicyHelper(helperConfig *S3PolicyHelperConfig) (*S3PolicyHelper, error) {
	var optFns []func(*config.LoadOptions) error
	if helperConfig != nil && helperConfig.AccessKeyID != "" && helperConfig.SecretAccessKey != "" {
		optFns = append(optFns, config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(helperConfig.AccessKeyID, helperConfig.SecretAccessKey, "")))
	}
	if helperConfig != nil && helperConfig.Region != "" {
		optFns = append(optFns, config.WithRegion(helperConfig.Region))
	}
	sdkConfig, err := config.LoadDefaultConfig(context.TODO(), optFns...)
	if err != nil {
		slog.Error("Failed to load AWS SDK config", "error", err)
		return nil, err
	}

	var s3OptFns []func(*s3.Options)
	if helperConfig != nil && helperConfig.Endpoint != "" {
		s3OptFns = append(s3OptFns, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(helperConfig.Endpoint)
		})
	}
	s3Client := s3.NewFromConfig(sdkConfig, s3OptFns...)
	return &S3PolicyHelper{
		client: s3Client,
		bucket: helperConfig.BucketName,
	}, nil
}

func (h *S3PolicyHelper) ClusterNodesUpdated(ctx context.Context, nodes []*interfaces.NodeInfo) error {
	var ips []string
	for _, node := range nodes {
		ips = append(ips, node.ExternalIPs...)
	}
	configData := configData{
		BucketName: h.bucket,
		IPs:        ips,
	}

	t := template.Must(template.New("bucketPolicy").Funcs(template.FuncMap{
		"sub": sub,
	}).Parse(bucketPolicyTemplate))
	var policyOutput strings.Builder
	if err := t.Execute(&policyOutput, configData); err != nil {
		slog.Error("Failed to generate S3 bucket policy", "error", err)
		return err
	}

	_, err := h.client.PutBucketPolicy(ctx, &s3.PutBucketPolicyInput{
		Bucket: aws.String(h.bucket),
		Policy: aws.String(policyOutput.String()),
	})
	if err != nil {
		slog.Error("Failed to update S3 bucket policy", "error", err)
		return err
	}
	return nil
}
