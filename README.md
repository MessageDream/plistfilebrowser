plistfilebrowser
================
此项目是一个ipa在线安装服务器，支持部署在各个平台，在第一次浏览目录的时候，程序会根据ipa文件生成plist文件,即为浏览的文件，客户端浏览器点击安装。
使用方法：
   1. go build server.go
   2. ./server https 8080 /ipadir (https或者http,8080为端口号,/ipadir是存放ipa文件的目录，必须是绝对路径)
