GOFMT=gofmt

.PHONY: test-%
test-%:
	@echo "Running $* tests..."
	gotestsum \
		--format standard-verbose \
		--rerun-fails=3 \
		--packages="./..." \
		--junitfile test-results/TEST-$*.xml

fmt:
	$(GOFMT) -w .
