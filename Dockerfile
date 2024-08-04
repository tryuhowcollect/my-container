FROM ubuntu:22.04

# 最小限のインストールを解除
RUN yes | unminimize

# 基本的なパッケージのインストール
RUN apt-get update && \
    apt-get install -y locales vim tmux less wget iproute2 iputils-ping

# タイムゾーンとロケールの設定
RUN DEBIAN_FRONTEND=noninteractive apt-get install -y tzdata 
RUN locale-gen ja_JP.UTF-8
ENV LANG=ja_JP.UTF-8
ENV TZ=Asia/Tokyo

# 作業ディレクトリの設定
WORKDIR /app

# Goのインストール
RUN wget https://golang.org/dl/go1.20.5.linux-arm64.tar.gz && \
    tar -C /usr/local -xzf go1.20.5.linux-arm64.tar.gz && \
    echo "export PATH=\$PATH:/usr/local/go/bin" >> ~/.profile && \
    echo "export GOPATH=\$HOME/go" >> ~/.profile && \
    /bin/bash -c "source ~/.profile"

# 環境変数を再設定（再ログインしなくても有効にするため）
ENV PATH="/usr/local/go/bin:${PATH}"
ENV GOPATH="/root/go"

# プロファイルの読み込み
RUN /bin/bash -c "source ~/.profile"
