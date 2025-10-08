# make/tools.mk
ROOT_DIR ?= $(abspath $(dir $(lastword $(MAKEFILE_LIST)))/..)
-include $(ROOT_DIR)/make/common.mk

.PHONY: dump-all-code
dump-all-code:
	@echo "→ Dumping all Go source files into $(ROOT_DIR)/all_code.txt"
	@> "$(ROOT_DIR)/all_code.txt"
	@find "$(ROOT_DIR)" -type f -name "*.go" \
	  -not -path "*/vendor/*" -not -path "*/.git/*" -print0 \
	  | xargs -0 -n 1 sh -c ' \
	      echo "================" >> "$(ROOT_DIR)/all_code.txt"; \
	      echo "// FILE: $$1"    >> "$(ROOT_DIR)/all_code.txt"; \
	      echo "================" >> "$(ROOT_DIR)/all_code.txt"; \
	      cat "$$1"               >> "$(ROOT_DIR)/all_code.txt"; \
	      echo >> "$(ROOT_DIR)/all_code.txt"; \
	      echo >> "$(ROOT_DIR)/all_code.txt"; \
	    ' sh
	@echo "✓ Written to $(ROOT_DIR)/all_code.txt"