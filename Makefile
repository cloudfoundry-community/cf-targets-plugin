.DEFAULT_GOAL := help

PROJECT        :=cf-targets-plugin
GOOS           :=$(shell go env GOOS)
GOARCH         :=$(shell go env GOARCH)
GOMODULECMD    :=main
RELEASE_ROOT   ?=releases
DEV_TEST_BUILD =./$(PROJECT)
TARGETS        ?=linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64
COVERAGE_DIR   := coverage
GOSEC_EXCLUDE  ?= G204,G304

ifneq ($(VERSION),)
VERSION_SPLIT:=$(subst ., ,$(VERSION))
  ifneq ($(words $(VERSION_SPLIT)),3)
    $(error VERSION does not 3 parts $(VERSION))
  endif
else
VERSION_TAG:=$(shell (git describe --tags --abbrev=0 2>/dev/null || echo 0.0.0) | sed -e "s/^v//")
VERSION_SPLIT:=$(subst ., ,$(VERSION_TAG))
  ifneq ($(words $(VERSION_SPLIT)),3)
    $(error VERSION_TAG  does not 3 parts |$(words $(VERSION_SPLIT))|$(VERSION_TAG)|$(VERSION_SPLIT)|)
  endif
  VERSION_SPLIT:=$(wordlist 1, 2, $(VERSION_SPLIT)) $(shell echo $$(($(word 3,$(VERSION_SPLIT))+1)))
endif

IS_NOT_NUMBER:=$(shell echo $(VERSION_SPLIT) | sed -e 's/[0123456789]//g')

ifneq ($(words $(IS_NOT_NUMBER)), 0)
  $(error The version string contain non-numeric characters)
endif

SEMVER_MAJOR    ?=$(word 1,$(VERSION_SPLIT))
SEMVER_MINOR    ?=$(word 2,$(VERSION_SPLIT))
SEMVER_PATCH    ?=$(word 3,$(VERSION_SPLIT))
SEMVER_PRERELEASE ?=
SEMVER_BUILDMETA  ?=
BUILD_DATE        :=$(shell date -u -Iseconds)
BUILD_VCS_URL     :=$(shell git config --get remote.origin.url)
BUILD_VCS_ID      :=$(shell git log -n 1 --date=iso-strict-local --format="%h")
BUILD_VCS_ID_DATE :=$(shell TZ=UTC0 git log -n 1 --date=iso-strict-local --format='%ad')

build: SEMVER_PRERELEASE := dev

GO_LDFLAGS = -X '$(GOMODULECMD).SemVerMajor=$(SEMVER_MAJOR)' \
	         -X '$(GOMODULECMD).SemVerMinor=$(SEMVER_MINOR)' \
	         -X '$(GOMODULECMD).SemVerPatch=$(SEMVER_PATCH)' \
	         -X '$(GOMODULECMD).SemVerPrerelease=$(SEMVER_PRERELEASE)' \
	         -X '$(GOMODULECMD).SemVerBuild=$(SEMVER_BUILDMETA)' \
	         -X '$(GOMODULECMD).BuildDate=$(BUILD_DATE)' \
	         -X '$(GOMODULECMD).BuildVcsUrl=$(BUILD_VCS_URL)' \
	         -X '$(GOMODULECMD).BuildVcsId=$(BUILD_VCS_ID)' \
		     -X '$(GOMODULECMD).BuildVcsIdDate=$(BUILD_VCS_ID_DATE)'

# The build meta data is added when the build is done
#
SEMVER_VERSION := $(if $(SEMVER_MAJOR),$(SEMVER_MAJOR),$(error Missing SEMVER_MAJOR))
SEMVER_VERSION := $(SEMVER_VERSION)$(if $(SEMVER_MINOR),.$(SEMVER_MINOR),$(error Missing SEMVER_MINOR))
SEMVER_VERSION := $(SEMVER_VERSION)$(if $(SEMVER_PATCH),.$(SEMVER_PATCH),$(error Missing SEMVER_PATCH))
SEMVER_VERSION := $(SEMVER_VERSION)$(if $(SEMVER_PRERELEASE),-$(SEMVER_PRERELEASE))

.PHONY: help
help: ## Display this help message
	@awk 'BEGIN {FS = ":.*##"; printf "Usage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[33m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Build

.PHONY: build install distbuild require-%

build: BUILD_GO_LDFLAGS:=-ldflags="$(GO_LDFLAGS) -X '$(GOMODULECMD).GoOs=$(GOOS)' -X '$(GOMODULECMD).GoArch=$(GOARCH)'"

build: BUILD_RULE_CMD := CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) \
	                     go build $(BUILD_GO_LDFLAGS) -o $(DEV_TEST_BUILD)

build: clean ## Build the binary for current platform
	@echo "Building $(DEV_TEST_BUILD)"
	$(BUILD_RULE_CMD)

install: build ## Build and install plugin into cf CLI
	cf install-plugin -f $(DEV_TEST_BUILD)

##@ Testing

.PHONY: test coverage coverage-html

test: ## Run all tests
	go test -race ./...

coverage: ## Generate test coverage report
	@mkdir -p $(COVERAGE_DIR)
	go test -race -coverprofile=$(COVERAGE_DIR)/coverage.out ./...
	go tool cover -func=$(COVERAGE_DIR)/coverage.out

coverage-html: coverage ## Generate and open HTML coverage report
	go tool cover -html=$(COVERAGE_DIR)/coverage.out -o $(COVERAGE_DIR)/coverage.html
	open $(COVERAGE_DIR)/coverage.html

##@ Code Quality

.PHONY: fmt vet lint lint-fix check

fmt: ## Format all Go source files
	gofmt -w .

vet: ## Run go vet
	go vet ./...

lint: ## Run golangci-lint
	@command -v golangci-lint >/dev/null 2>&1 || { \
		echo "golangci-lint not found. Install:"; \
		echo "  macOS:  brew install golangci-lint"; \
		echo "  Linux:  https://golangci-lint.run/welcome/install/"; \
		exit 1; \
	}
	golangci-lint run ./...

lint-fix: ## Run golangci-lint with auto-fix
	@command -v golangci-lint >/dev/null 2>&1 || { \
		echo "golangci-lint not found. Install:"; \
		echo "  macOS:  brew install golangci-lint"; \
		echo "  Linux:  https://golangci-lint.run/welcome/install/"; \
		exit 1; \
	}
	golangci-lint run --fix ./...

check: fmt vet lint ## Run all code quality checks (fmt, vet, lint)

##@ Dependencies

.PHONY: tidy

tidy: ## Tidy and verify go modules
	go mod tidy
	go mod verify

##@ Security

.PHONY: govulncheck trivy gosec gitleaks security

govulncheck: ## Run vulnerability check on dependencies
	@command -v govulncheck >/dev/null 2>&1 || { \
		echo "Installing govulncheck..."; \
		go install golang.org/x/vuln/cmd/govulncheck@latest; \
	}
	govulncheck ./...

trivy: ## Run Trivy filesystem vulnerability scanner
	@command -v trivy >/dev/null 2>&1 || { \
		echo "trivy not found. Install:"; \
		echo "  macOS:  brew install trivy"; \
		echo "  Linux:  https://aquasecurity.github.io/trivy/latest/getting-started/installation/"; \
		exit 1; \
	}
	trivy fs --scanners vuln,secret --severity HIGH,CRITICAL .

gosec: ## Run gosec security scanner (override GOSEC_EXCLUDE= to show all)
	@command -v gosec >/dev/null 2>&1 || { \
		echo "Installing gosec..."; \
		go install github.com/securego/gosec/v2/cmd/gosec@latest; \
	}
	gosec -quiet $(if $(GOSEC_EXCLUDE),-exclude=$(GOSEC_EXCLUDE)) ./...

gitleaks: ## Run gitleaks secret scanner
	@command -v gitleaks >/dev/null 2>&1 || { \
		echo "gitleaks not found. Install:"; \
		echo "  macOS:  brew install gitleaks"; \
		echo "  Linux:  https://github.com/gitleaks/gitleaks#installing"; \
		exit 1; \
	}
	gitleaks detect --source . --no-banner --redact

security: govulncheck trivy gosec gitleaks ## Run all security scans

##@ Verify

.PHONY: verify

verify: tidy check test security ## Run all checks before commit

##@ Release

.PHONY: ci-release release-all release-clean show-releases create-repo-index

require-%:
	@ if [ "${${*}}" = "" ]; then \
		echo "Environment variable $* not set"; \
		exit 1; \
	fi

RELEASES := $(foreach target,$(TARGETS),release-$(target)-$(PROJECT))

show-releases: ## Show built release artifacts
	@ls -lA $(RELEASE_ROOT)
	@echo ""

ci-release: require-VERSION release-all ## Build release (requires VERSION=x.y.z)

release-all: release-clean distbuild $(RELEASES) create-repo-index show-releases

create-repo-index: $(RELEASE_ROOT)/repo-index.yml

$(RELEASE_ROOT)/repo-index.yml: $(RELEASES) generate-repo-index
	./generate-repo-index "$(RELEASE_ROOT)" "$(PROJECT)" "$(SEMVER_VERSION)" "$(BUILD_DATE)"

distbuild:
	@mkdir -p $(RELEASE_ROOT)

define build-target
release-$(1)/$(2)-$(PROJECT): RELEASE_GO_LDFLAGS:=-ldflags="$(GO_LDFLAGS) -X '$(GOMODULECMD).GoOs=$(1)' -X '$(GOMODULECMD).GoArch=$(2)'"

release-$(1)/$(2)-$(PROJECT): RELEASE_EXECUTABLE_BASE:=$(RELEASE_ROOT)/$(PROJECT)-$(SEMVER_VERSION)+$(1).$(2)$(if $(3),.$(3))

release-$(1)/$(2)-$(PROJECT): RELEASE_EXECUTABLE:=$$(RELEASE_EXECUTABLE_BASE)$(if $(patsubst windows,,$(1)),,.exe)

release-$(1)/$(2)-$(PROJECT): RELEASE_EXECUTABLE_SHA1:=$$(RELEASE_EXECUTABLE_BASE).sha1

release-$(1)/$(2)-$(PROJECT):
	@echo "Building $$(PROJECT) version $$(SEMVER_VERSION) for $(1) $(2) ..."
	@CGO_ENABLED=0 GOOS=$(1) GOARCH=$(2) go build -o $$(RELEASE_EXECUTABLE) $$(RELEASE_GO_LDFLAGS)
	@openssl sha1 -r $$(RELEASE_EXECUTABLE) > $$(RELEASE_EXECUTABLE_SHA1)
endef

$(foreach target,$(TARGETS), $(eval $(call build-target,$(word 1, $(subst /, ,$(target))),$(word 2, $(subst /, ,$(target))),$(SEMVER_BUILDMETA))))

##@ Cleanup

.PHONY: clean release-clean distclean

clean: ## Clean dev build artifact
	@rm -f $(DEV_TEST_BUILD) || true

release-clean: ## Clean release artifacts
	@rm -f $(RELEASE_ROOT)/$(PROJECT)-* $(RELEASE_ROOT)/repo-index.yml || true
	@[[ ! -d $(RELEASE_ROOT) ]] || rmdir -p $(RELEASE_ROOT)

distclean: clean release-clean ## Clean all build artifacts
	rm -rf $(COVERAGE_DIR) coverage.out
