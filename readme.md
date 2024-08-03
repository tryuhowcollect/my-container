## コンテナの利点
Linuxコンテナを用いることで、独立した複数のLinux環境を稼働できる。<br>
コンテナは、ホストOSとリソースを共有するため軽量である。

## コンテナの正体について
以下のLinuxの機能を用いることで独立して動いているプロセスのこと
- Namespace
- Control Group
- File System

## 実行
コンテナ起動
```
docker compose up -d
```
コンテナ内に入る(Linux環境)
```
docker compose exec app bash
```
src配下で
```
go build -o my_container main.go
```
```
./my_container run /bin/bash
```

## 確認事項
- `syscall.CLONE_NEWUTS`を消すと、ホスト名の変更がホストシステム全体に及んでしまう<br>
→この名前空間のプロセスは独自のホスト名とドメインを持つ

- `syscall.CLONE_NEWPID`を消すと、コンテナ内のプロセスがホストのプロセスと連続する<br>
→この名前空間のプロセスは独自のPID空間を持つ

- `syscall.CLONE_NEWNS`を消すと、一時ファイルのマウントがホストにまで及んでしまう<br>
→この名前空間のプロセスは独自のマウント名前空間を持つ

これらの名前空間を使用したプロセスは、コンテナ機能を実現するための必要条件である。<br>
他にも、プロセスに対しネットワーク名前空間やcgroup等も必要。


## 参考文献
https://zenn.dev/bloomer/articles/5fd4e929fdb77a<br>
https://kaminashi-developer.hatenablog.jp/entry/dive-into-swamp-container-scratch
