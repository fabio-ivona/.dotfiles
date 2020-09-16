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
   
   a new ssh config file will be created, containing (and keeping track) of my personal ssh configurations

</details>


## Credits

Inspired by (and forked from) [Freek Murze](https://freek.dev)'s [dotfiles](https://github.com/freekmurze/dotfiles)
