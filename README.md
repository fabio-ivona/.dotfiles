# Linux shell tools

## Installation

```shell
cd ~
git clone --recursive git@gitlab.com:fabio.ivona/.dotfiles.git 
cd .dotfiles
bash bootstrap
```

## Features

<details>
   <summary><strong>Powerful shell</strong></summary>
   
   - zsh shell with custom configurations
   - oh-my-zsh ([documentation](https://ohmyz.sh))
   - zsh autosuggestions ([documentation](https://github.com/zsh-users/zsh-autosuggestions))
   - powerlevel10k theme [inserire link]
</details>

<details>
   <summary><strong>Global gitignore file</strong></summary>
   
   the installation script will create (and set as excludefile in git globals) a *.global-gitignore* file in your home directory which will add global gitignore rules:
   
   - .idea
   - node_modules
   - npm-debug.log
   - yarn-error.log
   - vendor
   - .env
   - wp-config.php

</details>

<details>
   <summary><strong>Personal fonts directory</strong></summary>
   
   a *.fonts* folder will be added to the home directory, containin some useful font
   
   - MeslogLGS (useful for a nice display of powerlevel10k zsh theme)

</details>

<details>
   <summary><strong>Personal .ssh/config file</strong></summary>
   
   during the installation process, the user is asked for his ssh config git repository, in order to clone it in the user ~/.dotfiles/shell/ssh/config folder and add a link to it from ~/.ssh/config
  
   this will allow the user to keep track of his personal ssh configurations

</details>


<details>
   <summary><strong>Tools</strong></summary>
   
   dotfiles will add its ~/.dotfiles/bin folder to PATH global variable, in order to add these scripts to the system toolbox:
   
   - [no scripts defined yet, will be added soon]

</details>



<details>
   <summary><strong>Commands</strong></summary>
   
   a number of aliases will be defined for the zsh shell:
   
   ###### PHP
  
   - `phpunit` executes phpunit tests from current directory (phpunit must be present composer.json file)
   - `dusk` executes dusk tests from current directory (dusk must be present composer.json file)
   - `artisan` executes artisan commands without the need to type *php artisan*
   - `deploy` executes laravel envoy deployment (*envoy run deploy*)
   - `deploy-code` executes laravel envoy deployment (*envoy run deploy-code*)
   
   ###### Misc
   - `sudo` allows to call sudo before aliases
   - `phpstorm` opens a PhpStorm project in current folder
   - `hostfile` opens a text editor for */etc/hosts* file 
   - `sshconfig` opens a text editor for *~/.ssh/config* file 
   - `dock` runs a *php dock* command (for dock info, see its [documentation](https://gitlab.com/defstudio/dock)) 
   
   ###### Git
   - `glog` show current project's git commits log in a readable way
   
   ###### Tools
   - `ll` shortcut for *ls -lF*
   - `l` shortcut for *ls -lF*
   - `la` shortcut for *ls -lFA*
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
   <summary><strong>Custom local dotfiles</strong></summary>
   
   along with default dotfiles (.aliases, .functions, .exports), user may add a ~/.dotfiles-custom/shell directory with additional .exports, .aliases, .functions, .zshrc files that will bel loaded after the default ones   
   
   these files will not be put under VCS

</details>


## Credits

Inspired by (and forked from) [Freek Murze](https://freek.dev)'s [dotfiles](https://github.com/freekmurze/dotfiles)
