#!/bin/bash

RESET="\033[0m"
GREEN="\033[32m"
YELLOW="\033[33m"
CYAN="\033[36m"

echo -e "${CYAN}--- hypr-dock & hypr-alttab Installer ---${RESET}"
echo ""

# Build first
echo -e "${YELLOW}Building binaries...${RESET}"
make build
if [ $? -ne 0 ]; then
    echo "Build failed. Exiting."
    exit 1
fi
echo -e "${GREEN}Build successful.${RESET}"
echo ""

# Ask about hypr-dock
read -p "Do you want to install hypr-dock (The Dock)? [Y/n] " install_dock
install_dock=${install_dock:-Y}

# Ask about hypr-alttab
read -p "Do you want to install hypr-alttab (The Alt-Tab Switcher)? [Y/n] " install_alttab
install_alttab=${install_alttab:-Y}

echo ""

if [[ "$install_dock" =~ ^[Yy]$ ]] && [[ "$install_alttab" =~ ^[Yy]$ ]]; then
    echo -e "${YELLOW}Installing BOTH hypr-dock and hypr-alttab...${RESET}"
    make install-all
elif [[ "$install_dock" =~ ^[Yy]$ ]]; then
    echo -e "${YELLOW}Installing ONLY hypr-dock...${RESET}"
    make install-dock
elif [[ "$install_alttab" =~ ^[Yy]$ ]]; then
    echo -e "${YELLOW}Installing ONLY hypr-alttab...${RESET}"
    make install-alttab
else
    echo "Nothing selected to install."
    exit 0
fi

echo ""
echo -e "${CYAN}Configuration files are located at: ~/.config/hypr-dock/${RESET}"
if [[ "$install_alttab" =~ ^[Yy]$ ]]; then
    echo -e "${YELLOW}IMPORTANT: Check ~/.config/hypr-dock/switcher.jsonc to configure hypr-alttab!${RESET}"
fi
echo -e "${GREEN}Done!${RESET}"
