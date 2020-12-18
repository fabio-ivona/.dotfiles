#!/bin/bash

echo ''
echo 'Installing phpstorm-url-handler script'
sudo rm -f /usr/bin/phpstorm-url-handler
sudo ln -s $HOME/.dotfiles/misc/phpstorm-url-handler/phpstorm-url-handler /usr/bin/phpstorm-url-handler
sudo chmod +x /usr/bin/phpstorm-url-handler

echo ''
echo 'Installing phpstorm-url-handler.desktop entry'
sudo desktop-file-install $HOME/.dotfiles/misc/phpstorm-url-handler/phpstorm-url-handler.desktop
sudo update-desktop-database

echo ''
echo 'Configuring phpstorm:// whitelist policy'
sudo mkdir -p /etc/opt/chrome/policies/managed
sudo mkdir -p /etc/opt/chrome/policies/recommended
sudo rm -f /etc/opt/chrome/policies/managed/whitelist_phpstorm_url_protocol.json
sudo ln -s $HOME/.dotfiles/misc/phpstorm-url-handler/whitelist_phpstorm_url_protocol.json /etc/opt/chrome/policies/managed/whitelist_phpstorm_url_protocol.json



