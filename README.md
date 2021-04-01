# pb_gin
pb_gin 是一个 protoc 的插件，可以通过 gRPC 的 service 定义文件和 googleapis/api/annotations 中
定义的 HttpRule 选项生成用于注册进 gin route 的代码。  
# 支持的特性
1. 自动解析请求参数进 proto 生成的结构。
2. 支持生成 swag 注释。
# 使用方法
## 安装 proto 插件
```bash
go get -u github.com/gu827356/pb-gin/cmd/protoc-gen-gin
```

## 在自己的工程中生成代码
```bash
protoc -I ${PROTO_IDL_DIR} --plugin=protoc-gen-gin --gin_out . --gin_opt=module=${YOUR_MODULE} xxxx.proto
```

# How To
## 如何从 context 中获取请求，响应和错误数据
pb-gin 提供了三个 API 获取 context 中的这些数据。  
```go
    gins.RequestInContext()
    gins.ResponseInContext()
    gins.ErrorInContext()
```
