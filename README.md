# Linux .dotfiles

This repository is a collection of scripts I use for a quick configuration of my develpment system.

I'm a fullstack web developer, so the tools I use are very opinionated. Feel free to use this respository as is or as a starting point for creating your own.

## Installation

```shell
cd ~
git clone --recursive git@github.com:fabio-ivona/.dotfiles.git 
cd .dotfiles
bash bootstrap
```

## Features

<details>
   <summary><strong>Powerful shell</strong></summary>
   
   - zsh shell with custom configurations
   - oh-my-zsh ([documentation](https://ohmyz.sh))
   - zsh autosuggestions ([documentation](https://github.com/zsh-users/zsh-autosuggestions))
   - darkula theme: enabled by default ([documentation](https://github.com/dracula/zsh))
   - powerlevel10k theme: not enabled by default ([documentation](https://github.com/romkatv/powerlevel10k))
</details>

<details>
   <summary><strong>Global gitignore file</strong></summary>
   
   the installation script will create (and set as excludefile in git globals) a *.global-gitignore* file in your home directory which will add global gitignore rules:
   
   - .idea
   - nohup.out
   - node_modules
   - npm-debug.log
   - yarn-error.log
   - vendor
   - .env
   - wp-config.php
   - error.log 
   - access.log

</details>

<details>
   <summary><strong>Personal fonts directory</strong></summary>
   
   a *.fonts* folder will be added to the home directory, containin some useful fonts
   
   - MeslogLGS (useful for a nice display of powerlevel10k zsh theme)

</details>

<details>
   <summary><strong>Personal .ssh/config file</strong></summary>
   
   during the installation process, you will be (optionally) asked for your ssh config git repository, in order to clone it in your ~`/.dotfiles/shell/ssh/config` folder and add a link to it from `~/.ssh/config`
  
   this will allow the user to keep track of your personal ssh configurations

</details>


<details>
   <summary><strong>Extra</strong></summary>
   
   dotfiles adds some extra system configuration:
   
   - larger bash history (32768 entries)
   - larger bash history file size (32768 entries)
   - lgnores duplicate commands in bash history
   - ignores commands which start with a space in bash history
   - ignores frequent commands both in history and in history file
      - ll
      - l
      - la
      - ls
      - cd
      - cd -
      - pwd
      - exit
      - date
      - --help commands

</details>



<details>
   <summary><strong>Commands</strong></summary>
   
   a number of aliases and functions will be defined for the zsh shell:
   
   ###### PHP
  
   - `phpunit` executes phpunit tests from current directory (phpunit must be present in composer.json file)
   - `pest` executes pestphp tests from current directory (pest must be present in composer.json file)
   - `dusk` executes dusk tests from current directory (dusk must be present in composer.json file)
   - `artisan` executes artisan commands without the need to type *php artisan*
   
   ###### Misc
   - `sudo` allows to call sudo before aliases
   - `phpstorm` opens a PhpStorm project in current folder
   - `hostfile` opens a text editor for */etc/hosts* file 
   - `sshconfig` opens a text editor for *~/.ssh/config* file 
   - `dock` runs a *php dock* command (for dock info, see its [documentation](https://github.com/def-studio/dock)) 
   
   ###### Git
   - `glog` show current project's git commits log in a readable way
   
   ###### Tools
   - `ll` shortcut for *ls -lF*
   - `l` shortcut for *ls -lF*
   - `la` shortcut for *ls -lFA*
   - `grep` colorize grep results
   - `publicip` shows current public IP
   - `localip` shows current local IPs
   - `mkd` creates a folder and move into it
   - `archive` create a zip archive of a folder
   

</details>


<details>
   <summary><strong>Custom local dotfiles</strong></summary>
   
   along with default dotfiles (.aliases, .functions, .exports), user may add a ~/.dotfiles-custom/shell directory with additional .exports, .aliases, .functions, .zshrc files that will bel loaded after the default ones   
   
   these files will not be put under VCS

</details>

<details>
   <summary><strong>Albert</strong></summary>

Albert (see [documentation](https://albertlauncher.github.io/)) is a useful launcher for ubuntu inspired by mac's Alfred. During the installation process you will be prompted to optionally install it

</details>

<details>
   <summary><strong>Dev Environment</strong></summary>

   a Web Development environment will be set up, with the following tools:

   - Php 8.0 cli (along with some extensions: xml, mbstring, intl)
   - Composer
   - Npm
   - Docker (and docker-compose)
   - PhpStorm
   - Sublime Text
   - Android Studio

</details>



<details>
   <summary><strong>PhpStorm URL Handler</strong></summary>
   
   dotfiles creates a .desktop entry to handle phpstorm://open?file=xxx links.

   NOTE: Laravel Ignition links may not work in a dockerized development environment. In this case, the local path should be mapped in Laravel's configuration. This can be done locally in .env file by adding this entry:

    IGNITION_LOCAL_SITES_PATH=/home/projects/example/src

</details>

<details>
   <summary><strong>Other Tools</strong></summary>

   some other everyday tools will be installed:
   
   - Thunderbird
   - Telegram Desktop
   - Whatsdesk
   - Libreoffice
   - htop
    
</details>


## Credits

Inspired by (and forked from) [Freek Murze](https://freek.dev)'s [dotfiles](https://github.com/freekmurze/dotfiles)
