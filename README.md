# AWS ALB Route Directive Adapter For Istio

AWS alb route directive adapter for istio is used in [Kubeflow](https://www.kubeflow.org/) [AWS Coginito manifest](https://github.com/kubeflow/manifests/blob/e55cf3d73c68c56752b7c51f4c4de01062eff610/kfdef/kfctl_aws_cognito.yaml#L92-L96) as part of the Authentication and Authorization offering.

**Authentication** - After your load balancer authenticates a user successfully, it sends the user claims received from the IdP to the target. The load balancer signs the user claim so that applications can verify the signature and verify that the claims were sent by the load balancer. The load balancer adds the HTTP headers `x-amzn-oidc-data` which is user claims, in JSON web tokens (JWT) format.

**Authorization** - Route directives enable Mixer adapters to modify traffic metadata using operation templates on the request and response headers. AWS alb route directive adapter for istio decode `x-amzn-oidc-data` and retrieve `email` field and add custom http header `kubeflow-userid: alice@amazon.com` which will be used by Kubeflow Authorization layer.

This repo is built from [Route directive adapter development guide](https://github.com/istio/istio/wiki/Route-directive-adapter-development-guide). If you meet any problems, please follow that instructions.


## Compability

- This is originally built with istio 1.1.x. In istio 1.2.x, Adapter was removed and you have to enable it manually. Check [change notes](https://istio.io/latest/news/releases/1.2.x/announcing-1.2/change-notes/)
- This is not limited to `Kubeflow` usage. You can use it for any similar use case and customize it.

## Development Instruction

### Istio Local Copy
If you don't have isito, please download nad build a local copy of istio

```
mkdir -p $GOPATH/src/istio.io/
cd $GOPATH/src/istio.io/
git clone https://github.com/istio/istio
cd istio
go build ./...
```

### Download code

Clone this repo and put it under `$GOPATH/src/istio.io/isito`.

### Build binary

```bash
cd $GOPATH/src/istio.io/istio
GO111MODULE=off CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o authzadaptor/authzadaptor ./authzadaptor/main/main.go
```

### Build container image

```bash
docker build -t seedjeffwan/istio-adapter:0.1 authzadaptor
```

### Regenerate codes

If you want to start from scratch, you can follow instruction to generate codes.

```bash
cd $GOPATH/src/istio.io/istio

bin/mixer_codegen.sh -t authzadaptor/template.proto
bin/mixer_codegen.sh -a authzadaptor/config/config.proto -x "-s=false -n authzadaptor -t authzadaptor”
```

## Security

See [CONTRIBUTING](CONTRIBUTING.md#security-issue-notifications) for more information.

## License

This project is licensed under the Apache-2.0 License.
