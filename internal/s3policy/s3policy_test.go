package s3policy

import (
	"html/template"
	"strings"
	"testing"
)

func TestBucketPolicyTemplate(t *testing.T) {
	tem := template.Must(template.New("bucketPolicy").Funcs(template.FuncMap{
		"sub": sub,
	}).Parse(bucketPolicyTemplate))

	configData := configData{
		BucketName: "hello",
		IPs:        []string{"127.0.0.1", "127.0.0.2"},
	}

	var policyOutput strings.Builder
	if err := tem.Execute(&policyOutput, configData); err != nil {
		t.Errorf("Failed to execute template: %v", err)
	}

	t.Log(policyOutput.String())
}
