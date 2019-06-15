# testtools
echo server(tcp, udp, http, https, dns, quic)

# install go
echo "export PATH=${PATH}:/opt/go/bin" >> /etc/profile
echo "export GOROOT=/opt/go" >> /etc/profile
mkdir -p /opt/go/gopath
echo "export GOPATH=/opt/go/gopath" >> /etc/profile

# install go tools
mkdir -p $GOPATH/src && cd !$
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