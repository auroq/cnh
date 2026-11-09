SHELL := bash
.SHELLFLAGS := -euo pipefail -c
.DEFAULT_GOAL := help

ROOTS      := bootstrap cluster platform
OP_TOKEN   := pass:cnh/op-service-account-token
STATE_USER := op:CNH/cnh-tfstate/username
STATE_PASS := op:CNH/cnh-tfstate/password

OP  := secret-run --env OP_SERVICE_ACCOUNT_TOKEN=$(OP_TOKEN) --
TF  := $(OP) secret-run --env AWS_ACCESS_KEY_ID=$(STATE_USER) --env AWS_SECRET_ACCESS_KEY=$(STATE_PASS) -- terraform

export KUBECONFIG := $(CURDIR)/.kube/config

ES_PASSWORD  ?= widedoor-dev
ES_PORT      ?= 19200
WIDEDOOR_URL ?= https://widedoor.cnh.auroq.com
export ES_PASSWORD ES_PORT

.PHONY: help
help: ## Show this help
	@grep -hE '^[a-zA-Z0-9_%-]+:.*## ' $(MAKEFILE_LIST) | awk -F':.*## ' '{printf "  %-22s %s\n", $$1, $$2}'
	@echo
	@echo "  ROOT is one of: $(ROOTS)   e.g. make plan-cluster"
	@echo "  KUBECONFIG is $(KUBECONFIG); it never touches ~/.kube/config"

.PHONY: bootstrap-seed
bootstrap-seed: ## First run only: create Spaces key + state bucket with local state, then move state into the bucket
	printf 'terraform {\n  backend "local" {}\n}\n' > infra/bootstrap/zz_seed_override.tf
	$(OP) terraform -chdir=infra/bootstrap init -upgrade
	$(OP) terraform -chdir=infra/bootstrap apply -target=onepassword_item.spaces_tfstate -target=digitalocean_project.cnh
	$(OP) terraform -chdir=infra/bootstrap apply
	rm infra/bootstrap/zz_seed_override.tf
	$(TF) -chdir=infra/bootstrap init -migrate-state -force-copy
	rm -f infra/bootstrap/terraform.tfstate infra/bootstrap/terraform.tfstate.backup

init-%: ## terraform init for ROOT
	$(TF) -chdir=infra/$* init -upgrade

validate-%: ## terraform validate for ROOT (no credentials needed)
	terraform -chdir=infra/$* init -backend=false -upgrade >/dev/null
	terraform -chdir=infra/$* validate

plan-%: ## terraform plan for ROOT
	$(TF) -chdir=infra/$* plan

apply-%: ## terraform apply for ROOT
	$(TF) -chdir=infra/$* apply

destroy-%: ## terraform destroy for ROOT
	$(TF) -chdir=infra/$* destroy

.PHONY: up
up: init-cluster apply-cluster kubeconfig init-platform apply-platform ## Cluster, kubeconfig, then platform add-ons

.PHONY: kubeconfig
kubeconfig: ## Write the cluster kubeconfig to .kube/config (mode 600)
	mkdir -p .kube
	umask 077; $(TF) -chdir=infra/cluster output -raw kubeconfig > .kube/config
	kubectl config rename-context "$$(kubectl config current-context)" cnh
	$(MAKE) --no-print-directory kubectx-merge
	kubectl get nodes -o wide

.PHONY: kubectx-merge
kubectx-merge: ## Merge the cnh context into ~/.kube/config (backs it up first) so kubectx/k9s see it
	cp ~/.kube/config ~/.kube/config.bak-cnh-$$(date +%Y%m%d%H%M%S)
	KUBECONFIG=$$HOME/.kube/config kubectl config delete-context cnh >/dev/null 2>&1 || true
	umask 077; KUBECONFIG=$$HOME/.kube/config:$(CURDIR)/.kube/config kubectl config view --flatten > $(CURDIR)/.kube/merged
	mv $(CURDIR)/.kube/merged ~/.kube/config
	@echo "cnh added to ~/.kube/config; default is still $$(KUBECONFIG=$$HOME/.kube/config kubectl config current-context). Run 'make use-cnh' to switch."

.PHONY: use-cnh
use-cnh: ## Make cnh the default context in ~/.kube/config for the event
	KUBECONFIG=$$HOME/.kube/config kubectl config use-context cnh

.PHONY: use-wasp
use-wasp: ## Switch the default context back to wasp
	KUBECONFIG=$$HOME/.kube/config kubectl config use-context lke195884-ctx

.PHONY: status
status: ## Nodes, non-running pods, gateways and certificates
	kubectl get nodes
	kubectl get pods -A --field-selector=status.phase!=Running,status.phase!=Succeeded
	kubectl get gateway,httproute -A
	kubectl get certificate -A

.PHONY: fmt
fmt: ## terraform fmt everything
	terraform fmt -recursive

.PHONY: teardown
teardown: ## Destroy platform, cluster, then the state bucket (moves bootstrap state back to local first)
	$(TF) -chdir=infra/platform destroy
	$(TF) -chdir=infra/cluster destroy
	printf 'terraform {\n  backend "local" {}\n}\n' > infra/bootstrap/zz_seed_override.tf
	$(TF) -chdir=infra/bootstrap init -migrate-state -force-copy
	$(OP) terraform -chdir=infra/bootstrap destroy
	rm infra/bootstrap/zz_seed_override.tf

.PHONY: test
test: ## Go unit tests, vet and gofmt check
	test -z "$$(gofmt -l .)"
	go vet ./...
	go test -race ./...

.PHONY: es-up
es-up: ## Start a local Elasticsearch (trial licence, ML on) with docker compose and wait until it and ELSER are ready
	docker compose up -d --wait es
	ES_URL=http://127.0.0.1:$(ES_PORT) tools/es-warm.sh

.PHONY: es-down
es-down: ## Stop the local Elasticsearch and delete its data
	docker compose down -v

.PHONY: test-integration
test-integration: es-up ## Integration tests against the local docker compose Elasticsearch
	ES_URL=http://127.0.0.1:$(ES_PORT) ES_USERNAME=elastic go test -race -count=1 ./testing/integration/...

.PHONY: test-acceptance
test-acceptance: ## Acceptance tests over HTTP against WIDEDOOR_URL (default: the deployed app)
	WIDEDOOR_URL=$(WIDEDOOR_URL) go test -race -count=1 ./testing/acceptance/...

.PHONY: test-all
test-all: test test-integration test-acceptance ## Unit, integration and acceptance tests

.PHONY: build
build: ## Build the Go binaries into bin/
	mkdir -p bin
	go build -o bin/ ./cmd/...

.PHONY: ingest
ingest: ## Run the schedule ingest against the cluster's Elasticsearch through a port-forward
	kubectl -n wide-door port-forward svc/wide-door-es-http 19200:9200 >/dev/null 2>&1 & pf=$$!; sleep 3; \
	secret-run --env WIDEDOOR_ELASTICSEARCH_PASSWORD=file:<(kubectl -n wide-door get secret wide-door-es-elastic-user -o jsonpath='{.data.elastic}' | base64 -d) -- WIDEDOOR_ELASTICSEARCH_URL=http://localhost:19200 go run ./cmd/widedoor ingest; \
	kill $$pf

.PHONY: issues
issues: ## Mirror roadmap items to GitHub issues (ADR 0003)
	tools/issues.sh
