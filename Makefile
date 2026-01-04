PROTO_SRC_DIR := proto
PROTO_OUT_DIR := build/proto

PROTO_FILES := $(shell find $(PROTO_SRC_DIR) -name "*.proto")

.PHONY: proto
proto: $(PROTO_FILES)
	@mkdir -p $(PROTO_OUT_DIR)
	protoc \
		--proto_path=$(PROTO_SRC_DIR) \
		--go_out=$(PROTO_OUT_DIR) \
		--go_opt=paths=source_relative \
		$?