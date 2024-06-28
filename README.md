# KubePoke

Helpers for managing Kubernetes node recycling with external resources.

## Features

KubePoke invokes helpers when nodes in a Kubernetes cluster change, such as:

- Automatically updating the HAProxy configuration with the internal IPs of the nodes and reloading HAProxy.
- Updating the bucket policy of S3 or S3-like object storage to allow requests from the external IPs of the nodes.

## Configuration

KubePoke supports reading from the default `KUBECONFIG` file and AWS-related environment variables. You also need to add a `config.yml` file in the current directory or `/etc/kubepoke`.

```yaml
cron: "*/5 * * * *" # The cron schedule for the cronjob.
helpers: # The list of helpers that are used.
  - haproxy
  - s3
haproxy:
  httpInternalPort: 8080 # The internal NodePort of the HTTP service.
  httpsInternalPort: 8443 # The internal NodePort of the HTTPS service.
s3:
  bucketName: kubepokemedia # The name of the S3 bucket.
  region: us-east-1 # The region of the S3 bucket.
  endpoint: http://minio:9000 # The endpoint of the S3 service, if you are not using AWS S3.
```

## License                                                                                                

KubePoke is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
