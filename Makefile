SYSTEM_CONFIG_DIR = /etc/hypr-dock

PROJECT_BIN_DIR = bin
PROJECT_CONFIG_DIR = configs/default

EXECUTABLE_DOCK = hypr-dock
EXECUTABLE_ALTTAB = hypr-alttab

CMD_DOCK = ./cmd/hypr-dock/.
CMD_ALTTAB = ./cmd/hypr-alttab/.

RESET := \033[0m
GREEN := \033[32m
YELLOW := \033[33m

warn:
	@if [ ! -f "$(PROJECT_BIN_DIR)/$(EXECUTABLE_DOCK)" ]; then \
		echo -e "$(YELLOW)The first build may take an extremely long time due to linking with gtk3...$(RESET)"; \
	fi

build-all:
	$(MAKE) build-dock
	$(MAKE) build-alttab

build: build-all

build-dock:
	$(MAKE) warn
	go build -v -o $(PROJECT_BIN_DIR)/$(EXECUTABLE_DOCK) $(CMD_DOCK)

build-alttab:
	$(MAKE) warn
	go build -v -o $(PROJECT_BIN_DIR)/$(EXECUTABLE_ALTTAB) $(CMD_ALTTAB)


update-dock:
	-sudo killall $(EXECUTABLE_DOCK) 2>/dev/null || true
	sudo cp $(PROJECT_BIN_DIR)/$(EXECUTABLE_DOCK) /usr/bin/
	@echo -e "$(GREEN)hypr-dock update.$(RESET)"

update-alttab:
	-sudo killall $(EXECUTABLE_DOCK) 2>/dev/null || true
	sudo cp $(PROJECT_BIN_DIR)/$(EXECUTABLE_DOCK) /usr/bin/
	@echo -e "$(GREEN)hypr-alttab update.$(RESET)"

install: install-all

install-dock:
	-sudo killall $(EXECUTABLE_DOCK) 2>/dev/null || true
	sudo cp $(PROJECT_BIN_DIR)/$(EXECUTABLE_DOCK) /usr/bin/
	sudo mkdir -p $(SYSTEM_CONFIG_DIR)
	sudo cp -r $(PROJECT_CONFIG_DIR)/* $(SYSTEM_CONFIG_DIR)/
	@echo -e "$(GREEN)hypr-dock installed. Configs copied to $(SYSTEM_CONFIG_DIR)$(RESET)"

install-alttab:
	-sudo killall $(EXECUTABLE_ALTTAB) 2>/dev/null || true
	sudo cp $(PROJECT_BIN_DIR)/$(EXECUTABLE_ALTTAB) /usr/bin/
	sudo mkdir -p $(SYSTEM_CONFIG_DIR)
	# Only copy configs if they don't exist to avoid overwriting user changes
	sudo cp -n $(PROJECT_CONFIG_DIR)/* $(SYSTEM_CONFIG_DIR)/ || true
	@echo -e "$(GREEN)hypr-alttab installed. Configs copied to $(SYSTEM_CONFIG_DIR)$(RESET)"

install-all:
	-sudo killall $(EXECUTABLE_DOCK) 2>/dev/null || true
	-sudo killall $(EXECUTABLE_ALTTAB) 2>/dev/null || true
	sudo cp $(PROJECT_BIN_DIR)/$(EXECUTABLE_DOCK) /usr/bin/
	sudo cp $(PROJECT_BIN_DIR)/$(EXECUTABLE_ALTTAB) /usr/bin/
	sudo mkdir -p $(SYSTEM_CONFIG_DIR)
	sudo cp -r $(PROJECT_CONFIG_DIR)/* $(SYSTEM_CONFIG_DIR)/
	@echo -e "$(GREEN)Both hypr-dock and hypr-alttab installed. Configs copied to $(SYSTEM_CONFIG_DIR)$(RESET)"

uninstall:
	sudo rm -f /usr/bin/$(EXECUTABLE_DOCK)
	sudo rm -f /usr/bin/$(EXECUTABLE_ALTTAB)
	sudo rm -rf $(SYSTEM_CONFIG_DIR)
	@echo -e "$(GREEN)Uninstalled.$(RESET)"

exec:
	./bin/hypr-dock -dev