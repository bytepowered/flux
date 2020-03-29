# PermissionFilter - 权限验证过滤器

PermissionFilter 依赖于JWT过滤器的验证结果，用于验证JWT的用户ID(subject)是否具有访问某API的权限。

**验证过程说明**

1. 根据JWT验证过滤器的Token数据，获取`jwt-subject-id`所指定的用户数据；
2. 根据用户访问权限的三元组`(subjectId, method, pattern)`，获取对应访问权限；访问权限的获取方式按配置参数，有如下方式：
    a. Http方式，通过配置指定的`upstrem-host, upstream-uri,upstream-method`，从指定Http API加载签名密钥；
    a. Dubbo方式，通过配置指定的`upstream-uri(dubbo interface),upstream-method`，从指定Dubbo接口加载签名密钥；
3. 根据返回权限结果，判定访问是否授权。

## 构成用户访问权限的三元组：

1. subjectId: 即JWT中标记用户的字段；
2. httpMethod: 当前访问API的http method；
3. httpPattern: 当前访问API的pattern，即后端服务注册到网关的path pattern，并非当前Http访问路径path；

## 过滤器配置

**示例: 基于Dubbo协议的配置**

```toml
[PermissionVerification]
disabled = false
upstream-protocol = "dubbo"
upstream-uri = "net.bytepowered.lingxiao.PermissionService"
upstream-method = "verify"
```

**示例: 基于Http协议的配置**

```toml
[PermissionVerification]
disabled = false
upstream-protocol = "http"
upstream-host = "http://acl.bytepowered-internal.io:8080"
upstream-uri = "/permission/api"
upstream-method = "POST"
```

### 参数说明

- `upstream-protocol` 后端服务支持权限校验的协议。支持: \[http, dubbo\]
- `upstream-host` 后端服务地址。在http协议下使用
- `upstream-uri` 后端服务地址：在dubbo协议中为 interface 路径；在http协议下，是完整URL地址；
- `upstream-method` 后端服务方法：在dubbo协议中为接口方法名；在http协议下，是Http方法名；

## 权限校验接口签名和参数

**Dubbo接口签名**

```java
net.bytepowered.lingxiao.PermissionService#verify(String subjectId, String method, String pattern) String
```

**Http接口参数**

```text
subjectId={subjectId}&method={accessHttpMethod}&pattern={accessPathPattern}
```

### 响应数据要求

返回字符串内容`success`表示指定用户SubjectId对（method, pattern）授权访问权限。其它字符表示没有权限。

注：Http接口须在`200 OK`状态下返回。

```text
success
```


