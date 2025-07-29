# Xray
只做备份
bash <(curl -Ls https://raw.githubusercontent.com/rickyryan/Xray/master/install.sh)


从 v2-ui 迁移
首先在安装了 v2-ui 的服务器上安装最新版 x-ui，然后使用以下命令进行迁移，将迁移本机 v2-ui 的所有 inbound 账号数据至 x-ui，面板设置和用户名密码不会迁移

迁移成功后请关闭 v2-ui并且重启 x-ui，否则 v2-ui 的 inbound 会与 x-ui 的 inbound 会产生端口冲突

x-ui v2-ui
