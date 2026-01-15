LOCAL_CONFIG_DIR = $(HOME)/.config/hypr-dock

PROJECT_BIN_DIR = bin
PROJECT_CONFIG_DIR = configs

EXECUTABLE_DOCK = hypr-dock
EXECUTABLE_ALTTAB = hypr-alttab

RESET := \033[0m
GREEN := \033[32m
YELLOW := \033[33m

build:
	@if [ ! -f "$(PROJECT_BIN_DIR)/$(EXECUTABLE_DOCK)" ]; then \
		echo -e "$(YELLOW)The first build may take an extremely long time due to linking with gtk3...$(RESET)"; \
	fi
	go build -v -o $(PROJECT_BIN_DIR)/$(EXECUTABLE_DOCK) ./main/.
	go build -v -o $(PROJECT_BIN_DIR)/$(EXECUTABLE_ALTTAB) ./main/.

install: install-all

install-dock:
	-sudo killall $(EXECUTABLE_DOCK) 2>/dev/null || true
	sudo cp $(PROJECT_BIN_DIR)/$(EXECUTABLE_DOCK) /usr/bin/
	mkdir -p $(LOCAL_CONFIG_DIR)
	cp -r $(PROJECT_CONFIG_DIR)/* $(LOCAL_CONFIG_DIR)/
	@echo -e "$(GREEN)hypr-dock installed.$(RESET)"

install-alttab:
	-sudo killall $(EXECUTABLE_ALTTAB) 2>/dev/null || true
	sudo cp $(PROJECT_BIN_DIR)/$(EXECUTABLE_ALTTAB) /usr/bin/
	mkdir -p $(LOCAL_CONFIG_DIR)
	# Only copy configs if they don't exist to avoid overwriting user changes, or force? 
	# User request says "check config file", so we should ensure it exists.
	cp -n $(PROJECT_CONFIG_DIR)/* $(LOCAL_CONFIG_DIR)/ || true
	@echo -e "$(GREEN)hypr-alttab installed.$(RESET)"

install-all:
	-sudo killall $(EXECUTABLE_DOCK) 2>/dev/null || true
	-sudo killall $(EXECUTABLE_ALTTAB) 2>/dev/null || true
	sudo cp $(PROJECT_BIN_DIR)/$(EXECUTABLE_DOCK) /usr/bin/
	sudo cp $(PROJECT_BIN_DIR)/$(EXECUTABLE_ALTTAB) /usr/bin/
	mkdir -p $(LOCAL_CONFIG_DIR)
	cp -r $(PROJECT_CONFIG_DIR)/* $(LOCAL_CONFIG_DIR)/
	@echo -e "$(GREEN)Both hypr-dock and hypr-alttab installed.$(RESET)"

uninstall:
	sudo rm -f /usr/bin/$(EXECUTABLE_DOCK)
	sudo rm -f /usr/bin/$(EXECUTABLE_ALTTAB)
	rm -rf $(LOCAL_CONFIG_DIR)
	@echo -e "$(GREEN)Uninstalled.$(RESET)"

exec:
	./bin/hypr-dock -dev
