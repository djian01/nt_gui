.DEFAULT_GOAL := build
.PHONY: setup build run test icons mac-app linux-package
setup build run test icons mac-app linux-package:
	$(MAKE) -C desktop $@
