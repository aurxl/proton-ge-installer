# Proton GE CLI Installer
installs [proton-ge-custom](https://github.com/GloriousEggroll/proton-ge-custom)

## Usage

### installing latest GE:
``` sh
proton-ge-installer
```
``` sh
proton-ge-installer latest
```
### installing specific version
``` sh
proton-ge-installer -v 8-14
```
``` sh
proton-ge-installer -v GE-Proton8-14
```
``` sh
proton-ge-installer GE-Proton8-14
```
### installing to special steam dir
``` sh
proton-ge-installer -v GE-Proton8-14 -d /home/neo/.steam
```
### force override existing install
```
proton-ge-installer -f latest
```

## install

add following line to `~/.bashrc` or `~/.zshrc`
``` sh
export PATH=$PATH:$HOME/go/bin
```

then: 
``` sh
go install github.com/aurxl/proton-ge-installer
```
