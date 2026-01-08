LOCAL_CONFIG_DIR = $(HOME)/.config/hypr-dock

PROJECT_BIN_DIR = bin
PROJECT_CONFIG_DIR = configs

EXECUTABLE = hypr-dock

RESET := \033[0m
GREEN := \033[32m
YELLOW := \033[33m

install:
		sudo cp $(PROJECT_BIN_DIR)/$(EXECUTABLE) /usr/bin/

		mkdir -p $(LOCAL_CONFIG_DIR)
		cp -r $(PROJECT_CONFIG_DIR)/* $(LOCAL_CONFIG_DIR)/

		@echo -e "$(GREEN)Installation completed.$(RESET)"

uninstall:
		sudo rm -f /usr/bin/$(EXECUTABLE)

		rm -rf $(LOCAL_CONFIG_DIR)

		@echo -e "$(GREEN)Installation removed.$(RESET)"

update:
		sudo rm -f /usr/bin/$(EXECUTABLE)
		sudo cp $(PROJECT_BIN_DIR)/$(EXECUTABLE) /usr/bin/

		@echo -e "$(GREEN)Updating completed.$(RESET)"

get:
		go mod tidy

build:
		@if [ ! -f "$(PROJECT_BIN_DIR)/$(EXECUTABLE)" ]; then \
			echo -e "$(YELLOW)The first build may take an extremely long time due to linking with gtk3...$(RESET)"; \
		fi
		go build -v -o bin/hypr-dock ./main/.

exec:
		./bin/hypr-dock -dev
