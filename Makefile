# from https://github.com/Mrs4s/MiraiGo/blob/master/Makefile
.PHONY: protoc-gen-golite-version clean install-protoc-plugin proto
.DEFAULT_GOAL := proto

PROTO_DIR=client/packets/pb
PROTO_OUTPUT_PATH=client/packets
PROTO_IMPORT_PATH=client/packets


PROTO_FILES := \
	$(PROTO_DIR)/action/*.proto \
	$(PROTO_DIR)/login/*.proto \
	$(PROTO_DIR)/message/*.proto \
	$(PROTO_DIR)/system/*.proto \
	$(PROTO_DIR)/service/*.proto \
	$(PROTO_DIR)/service/highway/*.proto \
	$(PROTO_DIR)/service/oidb/*.proto \
	$(PROTO_DIR)/*.proto


PROTOC_GEN_GOLITE_VERSION := \
	$(shell grep "github.com/RomiChan/protobuf" go.mod | awk -F v '{print "v"$$2}')


protoc-gen-golite-version:
	@echo "Use protoc-gen-golite version: $(PROTOC_GEN_GOLITE_VERSION)"

clean:
	find . -name "*.pb.go" | xargs rm -f

install-protoc-plugin: protoc-gen-golite-version
	go install github.com/RomiChan/protobuf/cmd/protoc-gen-golite@$(PROTOC_GEN_GOLITE_VERSION)

proto: install-protoc-plugin
	protoc --golite_out=$(PROTO_OUTPUT_PATH) --golite_opt=paths=source_relative -I=$(PROTO_IMPORT_PATH) $(PROTO_FILES)

fmt:
	go vet -stdmethods=false ./...

.EXPORT_ALL_VARIABLES:
GO111MODULE = on
