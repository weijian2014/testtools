
git clone https://github.com/lucas-clemente/quic-go.git
git checkout -b gquic
git branch --set-upstream-to=origin/gquic gquic
export BRANCH_NAME=`git symbolic-ref --short -q HEAD` && git fetch --all && git reset --hard origin/${BRANCH_NAME} && git pull


build tag:
    GOOS=linux;GOARCH=amd64;GO111MODULE=on