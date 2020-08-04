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