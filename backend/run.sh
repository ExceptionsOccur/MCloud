#!/bin/sh

# 启动 Nginx（后台运行）
nginx

# 启动后端（前台运行，容器退出时一同退出）
exec ./mcloud-server
