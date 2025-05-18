# 1. 注意事项

- fabric-orderer 镜像和 fabric-peer 镜像虽然也能够通过 cmd 可执行文件进行编译, 
但是用于编译镜像的资源文件以及 Docker file 位于 fabric/3.0.0/images 文件夹下, 因为它们是基于 fabric-3.0.0 而创建的。