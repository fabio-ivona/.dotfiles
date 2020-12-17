#!/bin/bash

sudo rm -f /usr/bin/phpstorm-url-handler
sudo ln -s $HOME/.dotfiles/misc/phpstorm-url-handler/phpstorm-url-handler /usr/bin/phpstorm-url-handler
sudo chmod +x /usr/bin/phpstorm-url-handler

sudo desktop-file-install $HOME/.dotfiles/misc/phpstorm-url-handler/phpstorm-url-handler.desktop

sudo update-desktop-database
