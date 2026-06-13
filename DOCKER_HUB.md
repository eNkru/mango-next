# Docker Hub 发布指南

### 构建镜像

```bash
docker build -t mango .
```

### 发布到 Docker Hub

1. **标记镜像**，使用你的 Docker Hub 用户名和仓库名：
   ```bash
   docker tag mango your-dockerhub-username/mango:latest
   ```
2. **登录 Docker Hub**：
   ```bash
   docker login
   ```
3. **推送镜像**：
   ```bash
   docker push your-dockerhub-username/mango:latest
   ```
