# JWTVerificationFilter - JWT验证过滤器

JWTVerificationFilter 通过读取Http请求的 Authorization解析并验证Token的有效性。

**验证过程说明**

1. 从Header、Query、FormParam等数据源读取JWT的原始Token数据；
2. 根据Token的`issuer(iss)、subject(sub)`字段，获取对应的Token签名密钥；加载密钥的方式按配置参数，有如下方式：
    a. Http方式，通过配置指定的`verification-uri, verification-method`，从指定Http API加载签名密钥；
    a. Dubbo方式，通过配置指定的`verification-uri(dubbo interface), verification-method`，从指定Dubbo接口加载签名密钥；
3. 使用获取的签名密钥，对Token进行验证。通过后，将Token的claims设置到`Context.Attributes`中，供后续组件使用；

**Token数据传递**

JWT验证通过后，JWTVerificationFilter会将Token对象，
以固定Key`jwt-claims`储存到Context.ScopedValue中，其内部数据类型为`map[string]interface{}`。
并且，同时将Token的关键数据单独设置到Context.Attributes中：

1. 固定Key`X-Jwt-Subject`， 值为Token.subject字段值；
1. 固定Key`X-Jwt-Issuer`， 值为Token.issuer字段值；
1. 固定Key`X-Jwt-Token`， 值为Token原始字符串；

后续的组件，可以根据上述固定Key从AttrValue来获取JWT的TokenClaim单项数据。
也可以使用固定Key`jwt-claims`从ScopedValue中获取Token的全部Claims数据。

## 过滤器配置

**示例: 基于Dubbo协议的配置**

```toml
# JWT验证配置
[JsonWebTokenVerification]
disable = false
type-id = "JwtVerificationFilter"
[JsonWebTokenVerification.InitConfig]
# 在Http请求中查找Token的Key定义。默认: header:Authorization。支持域：[query, form, path, header, attr]
jwt-lookup-token = "header:Authorization"
jwt-issuer-key = "iss"
jwt-subject-key = "sub"
verification-protocol = "DUBBO"
verification-uri = "net.bytepowered.lingxiao.JWTCertService"
verification-method = "getCertKey"
```

**示例: 基于Http协议的配置**

```toml
[HttpJsonWebTokenVerification]
disable = false
type-id = "JwtVerificationFilter"
[HttpJsonWebTokenVerification.InitConfig]
jwt-lookup-token = "Authorization"
jwt-issuer-key = "iss"
jwt-subject-key = "sub"
verification-proto = "HTTP"
verification-uri = "http://foo.bar.com:8080/jwt"
verification-method = "POST"
```

### 参数说明

- `verification-protocol` 后端服务支持权限校验的协议。支持: \[HTTP, DUBBO\]
- `verification-uri` 后端服务地址：在dubbo协议中为 interface 路径；在http协议下，是完整URL地址；
- `verification-method` 后端服务方法：在dubbo协议中为接口方法名；在http协议下，是Http方法名；
- `jwt-issuer-key` 用于识别JWT标识Issuer的字段，默认为JWT标准："iss"；
- `jwt-subject-key` 用于识别JWT标识用户的字段，默认为JWT标准："sub"；

### JWT Claims 传递方式

**Dubbo协议**

在后端协议为Dubbo中，将通过Dubbo.Attachments传输Context.AttrValues；

**Http协议**

在后端协议为Http中，将通过Header传递。即Context.AttrValues的键值对将被设置到Http.Header中。

## 权限校验接口签名和参数

**Dubbo接口签名**

```java
net.bytepowered.lingxiao.JWTCertService#getCertKey(String issuer, String subject, Map<Object, Object> claims) String
```

**Http接口参数**

```text
issuer={jwtIssuer}&subject={jwtSubject}&claims={JSON.string(jwtClaims)}
```

### 响应数据要求

返回字符串内容，内容为JWT验证密钥。

注：Http接口须在`200 OK`状态下返回。

```text
--- CERT KEY ---
```


