.DEFAULT_GOAL = build

DOCKER_PUBLISH_TAG = dwflynn/arb:0.0.1

tools/ko = tools/bin/ko
tools/bin/%: tools/src/%/go.mod tools/src/%/pin.go
	cd $(<D) && GOOS= GOARCH= go build -o $(abspath $@) $$(sed -En 's,^import "(.*)".*,\1,p' pin.go)

build: $(tools/ko)
	$(tools/ko) publish --local .
.PHONY: build

push: $(tools/ko)
	docker tag $$($(tools/ko) publish --local .) $(DOCKER_PUBLISH_TAG)
	docker push $(DOCKER_PUBLISH_TAG)
.PHONY: push
