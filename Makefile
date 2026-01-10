PROTO_OUT_DIR := build/proto
# all proto source directories under services/*/proto
PROTO_DIRS := $(shell find services -type d -name proto)
PROTO_FILES := $(shell find $(PROTO_DIRS) -name "*.proto")

.PHONY: proto
proto: $(PROTO_FILES)
	@mkdir -p $(PROTO_OUT_DIR)
	protoc \
		$(foreach dir,$(PROTO_DIRS),-I $(dir)) \
		--go_out=$(PROTO_OUT_DIR) \
		--go_opt=paths=source_relative \
		$(PROTO_FILES)