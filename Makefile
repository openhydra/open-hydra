# Makefile

REGISTRY ?= registry.cn-shanghai.aliyuncs.com/openhydra
TAG ?=

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

GOOS ?= $(shell go env GOHOSTOS)
GOARCH ?= $(shell go env GOARCH)
COMMIT_REF ?= $(shell git rev-parse --verify HEAD)
BOILERPLATE_DIR = $(shell pwd)/hack
GOPROXY ?= https://goproxy.cn,direct
GOVERSION ?= 1.21.0
IMAGETAG ?= $(shell git rev-parse --abbrev-ref HEAD)-$(shell git rev-parse --verify HEAD)-$(shell date -u '+%Y%m%d%I%M%S')

APP ?=
ifeq ($(APP),)
apps = $(shell ls cmd)
else
apps = $(APP)
endif

RACE ?=
ifeq ($(RACE),on)
	race = "-race"
endif

TAG ?=
ifeq ($(TAG),)
	TAG = $(COMMIT_REF)
endif

.PHONY: update-openapi
update-openapi:
	$(GOBIN)/openapi-gen --input-dirs open-hydra/pkg/open-hydra/apis,open-hydra/pkg/apis/open-hydra-api/course/core/v1,open-hydra/pkg/apis/open-hydra-api/setting/core/v1,open-hydra/pkg/apis/open-hydra-api/summary/core/v1,open-hydra/pkg/apis/open-hydra-api/device/core/v1,open-hydra/pkg/apis/open-hydra-api/user/core/v1,open-hydra/pkg/apis/open-hydra-api/dataset/core/v1,k8s.io/apimachinery/pkg/util/intstr,k8s.io/apimachinery/pkg/api/resource,k8s.io/apimachinery/pkg/apis/meta/v1,k8s.io/apimachinery/pkg/runtime,k8s.io/api/core/v1,k8s.io/apimachinery/pkg/apis/meta/v1 \
	--output-package open-hydra/pkg/generated/apis/openapi --output-base ./..  --go-header-file $(BOILERPLATE_DIR)/boilerplate.go.txt

.PHONY: gen-device-deepcopy-set
gen-device-deepcopy-set:
	$(GOBIN)/deepcopy-gen --input-dirs open-hydra/pkg/apis/open-hydra-api/device/core/v1 --output-package  open-hydra/pkg/apis/open-hydra-api/device/core/v1 --output-base ./..  -O zz_generated.deepcopy --go-header-file  $(BOILERPLATE_DIR)/boilerplate.go.txt
	$(GOBIN)/register-gen --input-dirs open-hydra/pkg/apis/open-hydra-api/device/core/v1 --output-package  open-hydra/pkg/apis/open-hydra-api/device/core/v1 --output-base ./.. -O register  --go-header-file  $(BOILERPLATE_DIR)/boilerplate.go.txt

.PHONY: gen-dataset-deepcopy-set
gen-dataset-deepcopy-set:
	$(GOBIN)/deepcopy-gen --input-dirs open-hydra/pkg/apis/open-hydra-api/dataset/core/v1 --output-package  open-hydra/pkg/apis/open-hydra-api/dataset/core/v1 --output-base ./..  -O zz_generated.deepcopy --go-header-file  $(BOILERPLATE_DIR)/boilerplate.go.txt
	$(GOBIN)/register-gen --input-dirs open-hydra/pkg/apis/open-hydra-api/dataset/core/v1 --output-package  open-hydra/pkg/apis/open-hydra-api/dataset/core/v1 --output-base ./.. -O register  --go-header-file  $(BOILERPLATE_DIR)/boilerplate.go.txt

.PHONY: gen-course-deepcopy-set
gen-course-deepcopy-set:
	$(GOBIN)/deepcopy-gen --input-dirs open-hydra/pkg/apis/open-hydra-api/course/core/v1 --output-package  open-hydra/pkg/apis/open-hydra-api/course/core/v1 --output-base ./..  -O zz_generated.deepcopy --go-header-file  $(BOILERPLATE_DIR)/boilerplate.go.txt
	$(GOBIN)/register-gen --input-dirs open-hydra/pkg/apis/open-hydra-api/course/core/v1 --output-package  open-hydra/pkg/apis/open-hydra-api/course/core/v1 --output-base ./.. -O register  --go-header-file  $(BOILERPLATE_DIR)/boilerplate.go.txt

.PHONY: gen-user-deepcopy-set
gen-user-deepcopy-set:
	$(GOBIN)/deepcopy-gen --input-dirs open-hydra/pkg/apis/open-hydra-api/user/core/v1 --output-package  open-hydra/pkg/apis/open-hydra-api/user/core/v1 --output-base ./..  -O zz_generated.deepcopy --go-header-file  $(BOILERPLATE_DIR)/boilerplate.go.txt
	$(GOBIN)/register-gen --input-dirs open-hydra/pkg/apis/open-hydra-api/user/core/v1 --output-package  open-hydra/pkg/apis/open-hydra-api/user/core/v1 --output-base ./.. -O register  --go-header-file  $(BOILERPLATE_DIR)/boilerplate.go.txt

.PHONY: gen-setting-deepcopy-set
gen-setting-deepcopy-set:
	$(GOBIN)/deepcopy-gen --input-dirs open-hydra/pkg/apis/open-hydra-api/setting/core/v1 --output-package  open-hydra/pkg/apis/open-hydra-api/setting/core/v1 --output-base ./..  -O zz_generated.deepcopy --go-header-file  $(BOILERPLATE_DIR)/boilerplate.go.txt
	$(GOBIN)/register-gen --input-dirs open-hydra/pkg/apis/open-hydra-api/setting/core/v1 --output-package  open-hydra/pkg/apis/open-hydra-api/setting/core/v1 --output-base ./.. -O register  --go-header-file  $(BOILERPLATE_DIR)/boilerplate.go.txt

.PHONY: gen-summary-deepcopy-set
gen-summary-deepcopy-set:
	$(GOBIN)/deepcopy-gen --input-dirs open-hydra/pkg/apis/open-hydra-api/summary/core/v1 --output-package  open-hydra/pkg/apis/open-hydra-api/summary/core/v1 --output-base ./..  -O zz_generated.deepcopy --go-header-file  $(BOILERPLATE_DIR)/boilerplate.go.txt
	$(GOBIN)/register-gen --input-dirs open-hydra/pkg/apis/open-hydra-api/summary/core/v1 --output-package  open-hydra/pkg/apis/open-hydra-api/summary/core/v1 --output-base ./.. -O register  --go-header-file  $(BOILERPLATE_DIR)/boilerplate.go.txt

.PHONY: gen-all-deepcopy-set
gen-all-deepcopy-set: gen-device-deepcopy-set gen-dataset-deepcopy-set gen-user-deepcopy-set gen-summary-deepcopy-set gen-setting-deepcopy-set gen-course-deepcopy-set

.PHONY: test-all
test-all:
	ginkgo -r -v --cover --coverprofile=coverage.out

.PHONY: fmt
fmt:
	gofmt -w pkg cmd

.PHONY: vet
vet:
	go vet ./...

.PHONY: go-build
go-build:
	CGO_ENABLED=0 GOARCH=$(GOARCH) go build -o cmd/open-hydra-server/open-hydra-server -ldflags "-X 'main.version=${TAG}'" cmd/open-hydra-server/main.go

.PHONY: image
image: go-build
	docker build -t $(REGISTRY)/open-hydra-server:$(IMAGETAG) --load .

.PHONY: image-no-container
image-no-container: go-build
	docker build -f hack/builder/Dockerfile -t $(REGISTRY)/open-hydra-server:$(IMAGETAG) --load .
