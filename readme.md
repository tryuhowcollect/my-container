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

## 参考文献
https://zenn.dev/bloomer/articles/5fd4e929fdb77a<br>
https://kaminashi-developer.hatenablog.jp/entry/dive-into-swamp-container-scratch
