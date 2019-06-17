# testtools
echo server(tcp, udp, http, https, dns, quic)

# install go
echo "export PATH=${PATH}:/opt/go/bin" >> /etc/profile
echo "export GOROOT=/opt/go" >> /etc/profile
mkdir -p /opt/go/gopath
echo "export GOPATH=/opt/go/gopath" >> /etc/profile

# install clang8.0
#!/bin/sh
CURRENT_DIR=`pwd`
mkdir clang
cd clang
wget http://releases.llvm.org/8.0.0/llvm-8.0.0.src.tar.xz
tar -xf llvm-8.0.0.src.tar.xz
mv llvm-8.0.0.src llvm
cd llvm/tools/
wget http://releases.llvm.org/8.0.0/cfe-8.0.0.src.tar.xz
tar -xf cfe-8.0.0.src.tar.xz
mv cfe-8.0.0.src clang
cd ../projects
wget http://releases.llvm.org/8.0.0/compiler-rt-8.0.0.src.tar.xz
tar -xf compiler-rt-8.0.0.src.tar.xz
mv compiler-rt-8.0.0.src compiler-rt
mkdir -p /tmp/clang_build
cd /tmp/clang_build
cmake -G "Unix Makefiles" ${CURRENT_DIR}/clang/llvm
make -j 8
make instal

# install go tools
mkdir -p $GOPATH/src && cd $GOPATH/src
git clone https://github.com/golang/tools.git golang.org/x/tools
git clone https://github.com/golang/lint.git golang.org/x/lint
git clone https://github.com/golang/net.git golang.org/x/net
git clone https://github.com/golang/sys.git golang.org/x/sys
git clone https://github.com/golang/crypto.git golang.org/x/crypto
git clone https://github.com/golang/text.git golang.org/x/text
git clone https://github.com/golang/image.git golang.org/x/image
git clone https://github.com/golang/oauth2.git golang.org/x/oauth2

# settings.json
{
    "editor.fontSize": 16,
    "workbench.startupEditor": "newUntitledFile",
    "remote.SSH.showLoginTerminal": true,
    "go.goroot": "/opt/go",
    "go.gopath": "/opt/go/gopath",
    "go.buildOnSave": "workspace",
    "go.lintOnSave": "workspace",
    "go.vetOnSave": "workspace",
    "go.buildTags": "",
    "go.buildFlags": [],
    "go.lintFlags": [],
    "go.vetFlags": [],
    "go.coverOnSave": false,
    "go.useCodeSnippetsOnFunctionSuggest": false,
    "go.formatTool": "goreturns",
}

# keybindings.json
[
    {
        "key": "f7",
        "command": "workbench.action.tasks.runTask",
        "args": "go_build_server",
        "when": "editorTextFocus"
    },
    {
        "key": "f8",
        "command": "workbench.action.tasks.runTask",
        "args": "go_build_client",
        "when": "editorTextFocus"
    }
]