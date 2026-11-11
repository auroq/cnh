SHELL := bash
.SHELLFLAGS := -euo pipefail -c
.DEFAULT_GOAL := help

ROOTS      := bootstrap cluster platform node node-platform
NODE_KUBECONFIG := $(CURDIR)/.kube/node.config
CLUSTER_KUBECONFIG := $(CURDIR)/.kube/config
WASP_CONTEXT := lke195884-ctx
TF_ARGS    ?=
OP_TOKEN   := pass:cnh/op-service-account-token
STATE_USER := op:CNH/cnh-tfstate/username
STATE_PASS := op:CNH/cnh-tfstate/password

OP  := secret-run --env OP_SERVICE_ACCOUNT_TOKEN=$(OP_TOKEN) --
TF  := $(OP) secret-run --env AWS_ACCESS_KEY_ID=$(STATE_USER) --env AWS_SECRET_ACCESS_KEY=$(STATE_PASS) -- terraform

export KUBECONFIG := $(NODE_KUBECONFIG)

ES_PASSWORD  ?= widedoor-dev
ES_PORT      ?= 19200
WIDEDOOR_URL ?= https://widedoor.cnh.auroq.com
export ES_PASSWORD ES_PORT

.PHONY: help
help: ## Show this help
	@grep -hE '^[a-zA-Z0-9_%-]+:.*## ' $(MAKEFILE_LIST) | awk -F':.*## ' '{printf "  %-22s %s\n", $$1, $$2}'
	@echo
	@echo "  ROOT is one of: $(ROOTS)   e.g. make plan-cluster"
	@echo "  KUBECONFIG is $(KUBECONFIG) (the node); the event cluster targets use $(CLUSTER_KUBECONFIG). Only kubectx-* and use-* edit ~/.kube/config"

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

planfile-%: ## terraform plan for ROOT saved to infra/ROOT/ROOT.tfplan (pass TF_ARGS='-var x=y' for variables)
	$(TF) -chdir=infra/$* plan -out=$*.tfplan $(TF_ARGS)

applyfile-%: ## terraform apply of the saved infra/ROOT/ROOT.tfplan, nothing else
	$(TF) -chdir=infra/$* apply $*.tfplan
	rm -f infra/$*/$*.tfplan

.PHONY: up-node
up-node: init-node applyfile-node node-kubeconfig init-node-platform planfile-node-platform applyfile-node-platform ## Droplet, kubeconfig, then Argo CD and secrets on it (run planfile-node first and read it)

.PHONY: node-ip
node-ip: ## Print the droplet's public IPv4
	@$(TF) -chdir=infra/node output -raw ipv4

.PHONY: node-ssh
node-ssh: ## SSH to the droplet as root
	ssh root@$$($(TF) -chdir=infra/node output -raw ipv4)

.PHONY: node-kubeconfig
node-kubeconfig: ## Wait for cloud-init, then write the MicroK8s admin kubeconfig to .kube/node.config (mode 600)
	mkdir -p .kube
	ip=$$($(TF) -chdir=infra/node output -raw ipv4); \
	for i in $$(seq 1 60); do \
	  if ssh -o StrictHostKeyChecking=accept-new -o ConnectTimeout=5 root@$$ip test -e /var/lib/microk8s-bootstrap.done 2>/dev/null; then break; fi; \
	  echo "waiting for microk8s bootstrap on $$ip ($$i/60)"; sleep 10; \
	done; \
	umask 077; ssh root@$$ip microk8s config \
	  | sed -e "s#https://[0-9.]*:16443#https://$$ip:16443#" \
	        -e 's#^\(\s*\(- \)\?name:\|\s*cluster:\|\s*user:\|current-context:\) microk8s-cluster$$#\1 cnh-node#' \
	        -e 's#^\(\s*\(- \)\?name:\|\s*user:\) admin$$#\1 cnh-node-admin#' \
	        -e 's#^\(\s*\(- \)\?name:\|current-context:\) microk8s$$#\1 cnh-node#' \
	  > $(NODE_KUBECONFIG)
	kubectl get nodes -o wide
	$(MAKE) --no-print-directory kubectx-merge-node

.PHONY: node-bootstrap
node-bootstrap: ## Kibana token, ELSER download and the first ingest on the node
	KUBECONFIG=$(NODE_KUBECONFIG) tools/node-bootstrap.sh

.PHONY: node-status
node-status: ## Pods, ingress and certificate on the node
	KUBECONFIG=$(NODE_KUBECONFIG) kubectl get nodes
	KUBECONFIG=$(NODE_KUBECONFIG) kubectl -n wide-door get pods,ingress,certificate
	KUBECONFIG=$(NODE_KUBECONFIG) kubectl -n argocd get applications

.PHONY: up
up: init-cluster apply-cluster kubeconfig init-platform apply-platform ## Cluster, kubeconfig, then platform add-ons

.PHONY: kubeconfig
kubeconfig: ## Write the cluster kubeconfig to .kube/config (mode 600)
	mkdir -p .kube
	umask 077; $(TF) -chdir=infra/cluster output -raw kubeconfig > $(CLUSTER_KUBECONFIG)
	KUBECONFIG=$(CLUSTER_KUBECONFIG) kubectl config rename-context "$$(KUBECONFIG=$(CLUSTER_KUBECONFIG) kubectl config current-context)" cnh
	$(MAKE) --no-print-directory kubectx-merge
	KUBECONFIG=$(CLUSTER_KUBECONFIG) kubectl get nodes -o wide

.PHONY: kubectx-merge
kubectx-merge: ## Merge the cnh context into ~/.kube/config (backs it up first) so kubectx/k9s see it
	cp ~/.kube/config ~/.kube/config.bak-cnh-$$(date +%Y%m%d%H%M%S)
	KUBECONFIG=$$HOME/.kube/config kubectl config delete-context cnh >/dev/null 2>&1 || true
	umask 077; KUBECONFIG=$$HOME/.kube/config:$(CLUSTER_KUBECONFIG) kubectl config view --flatten > $(CURDIR)/.kube/merged
	mv $(CURDIR)/.kube/merged ~/.kube/config
	@echo "cnh added to ~/.kube/config; default is still $$(KUBECONFIG=$$HOME/.kube/config kubectl config current-context). Run 'make use-cnh' to switch."

.PHONY: kubectx-merge-node
kubectx-merge-node: ## Swap the node context in for the deleted cluster's in ~/.kube/config (backs it up first); the default context is unchanged
	cp ~/.kube/config ~/.kube/config.bak-cnh-$$(date +%Y%m%d%H%M%S)
	for c in cnh cnh-node; do KUBECONFIG=$$HOME/.kube/config kubectl config delete-context $$c >/dev/null 2>&1 || true; done
	KUBECONFIG=$$HOME/.kube/config kubectl config delete-cluster do-sfo3-cnh >/dev/null 2>&1 || true
	KUBECONFIG=$$HOME/.kube/config kubectl config delete-user do-sfo3-cnh-admin >/dev/null 2>&1 || true
	umask 077; KUBECONFIG=$$HOME/.kube/config:$(NODE_KUBECONFIG) kubectl config view --flatten > $(CURDIR)/.kube/merged
	mv $(CURDIR)/.kube/merged ~/.kube/config
	@echo "cnh-node is in ~/.kube/config; default is still $$(KUBECONFIG=$$HOME/.kube/config kubectl config current-context)"

.PHONY: use-cnh
use-cnh: ## Make cnh the default context in ~/.kube/config for the event
	KUBECONFIG=$$HOME/.kube/config kubectl config use-context cnh

.PHONY: use-wasp
use-wasp: ## Switch the default context back to wasp
	KUBECONFIG=$$HOME/.kube/config kubectl config use-context $(WASP_CONTEXT)

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

E2E_DIR := testing/e2e
export PLAYWRIGHT_BROWSERS_PATH := $(CURDIR)/$(E2E_DIR)/.cache/ms-playwright

$(E2E_DIR)/node_modules: $(E2E_DIR)/package-lock.json
	cd $(E2E_DIR) && npm ci --no-fund --no-audit
	touch $@

.PHONY: test-e2e
test-e2e: $(E2E_DIR)/node_modules ## Browser accessibility tests (Playwright + axe) against WIDEDOOR_URL; first run installs Chromium into testing/e2e/.cache
	cd $(E2E_DIR) && npx playwright install chromium
	cd $(E2E_DIR) && WIDEDOOR_URL=$(WIDEDOOR_URL) npx playwright test

.PHONY: test-all
test-all: test test-integration test-acceptance test-e2e ## Unit, integration, acceptance and browser tests

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

.PHONY: dashboard
dashboard: ## Push deploy/dash0/*.json to Dash0 (PUT by origin, token via secret-run)
	tools/dash0-dashboard.sh

.PHONY: dashboard-check
dashboard-check: ## Read back the Dash0 dashboards deploy/dash0/*.json describe
	tools/dash0-dashboard.sh --check
