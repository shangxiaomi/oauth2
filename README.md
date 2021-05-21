为了满足自己的需要，本项目基于[https://github.com/llaoj/oauth2](https://github.com/llaoj/oauth2)二次开发
# oauth2
## 局限性
在传统的C/S身份验证模型中，客户端通过使用资源所有者的凭据，向服务器进行身份验证，从而请求服务器访问受限的资源（受保护资源）。 为了使第三方应用程序能够访问受限资源，资源所有者需要与第三方共享其凭据。 这会产生一些问题和局限：
- 为了将来的需要，第三方程序需要存储资源拥有者的凭据（通常为明文密码）
- 即使密码验证存在安全漏洞，服务器仍然需要使用它
- 第三方程序对资源拥有者的受保护资源拥有过于宽泛的权限，同时资源拥有者也没有能力对第三方程序进行限制（如限制第三方程序仅访问部分资源，或限制第三方程序的访问时间等）
- 资源拥有者必须通过更改密码来撤销第三方应用的权限，并且不能对单个第三方应用撤权（一旦更改密码，所有之前授予权限的第三方应用程序都要重新授权）
- 任意第三方应用的泄密都会导致用户的密码和受该密码保护的所有数据泄密。 
  
## 解决方案
> OAuth通过引入授权层，将客户端的角色与资源所有者的角色分开来解决这些问题。客户端通过获取访问令牌而不是使用资源拥有者的凭据来访问受限资源。访问令牌是一个表示特定范围，生命周期和其他访问属性的字符串。授权服务器在资源所有者的批准下才会向第三方客户端颁发访问令牌。客户端使用访问令牌来访问资源服务器托管的受保护资源。
使用方法

## 四种角色
OAuth定义了四个角色
- 资源所有者（resource owner）：能够对受保护资源授予访问权限的实体。当资源所有者是一个人时，它被称为用户。
- 资源服务器（resource server）：托管受保护资源的服务器，能够接受和响应通过令牌对受保护的资源的请求。
- 客户端（client）：代表资源所有者及其授权进行受保护资源请求的应用程序。术语"客户端"并不按时任何特定的实现特征。
- 授权服务器（authorization server）：服务器向客户端发出令牌验证资源所有者并获得授权。


## 基础概念
- 访问令牌（access token）-- 访问受限资源的凭据
- 刷新令牌（refresh token）-- 获取access token的凭据
- 回调地址（redirect uri）
- 权限范围（scope）
1. 启动redis（如果使用内存seesion）可以忽略


## 运行流程
```
+--------+                               +---------------+
|        |--(A)- Authorization Request ->|   Resource    |
|        |                               |     Owner     |
|        |<-(B)-- Authorization Grant ---|               |
|        |                               +---------------+
|        |
|        |                               +---------------+
|        |--(C)-- Authorization Grant -->| Authorization |
| Client |                               |     Server    |
|        |<-(D)----- Access Token -------|               |
|        |                               +---------------+
|        |
|        |                               +---------------+
|        |--(E)----- Access Token ------>|    Resource   |
|        |                               |     Server    |
|        |<-(F)--- Protected Resource ---|               |
+--------+                               +---------------+
```
（A）用户打开客户端以后，客户端要求用户给予授权。

（B）用户同意给予客户端授权。

（C）客户端使用上一步获得的授权，向认证服务器申请令牌。

（D）认证服务器对客户端进行认证以后，确认无误，同意发放令牌。

（E）客户端使用令牌，向资源服务器申请获取资源。

（F）资源服务器确认令牌无误，同意向客户端开放资源。

## 四种授权模式
### 授权码模式（authorization code）
```
+----------+
| Resource |
|   Owner  |
|          |
+----------+
^
|
(B)
+----|-----+          Client Identifier      +---------------+
|         -+----(A)-- & Redirection URI ---->|               |
|  User-   |                                 | Authorization |
|  Agent  -+----(B)-- User authenticates --->|     Server    |
|          |                                 |               |
|         -+----(C)-- Authorization Code ---<|               |
+-|----|---+                                 +---------------+
|    |                                         ^      v
(A)  (C)                                        |      |
|    |                                         |      |
^    v                                         |      |
+---------+                                      |      |
|         |>---(D)-- Authorization Code ---------'      |
|  Client |          & Redirection URI                  |
|         |                                             |
|         |<---(E)----- Access Token -------------------'
+---------+       (w/ Optional Refresh Token)

Note: The lines illustrating steps (A), (B), and (C) are broken into
two parts as they pass through the user-agent.
```
（A）用户访问客户端，后者将前者导向认证服务器。

（B）用户选择是否给予客户端授权。

（C）假设用户给予授权，认证服务器将用户导向客户端事先指定的"重定向URI"（redirection URI），同时附上一个授权码。

（D）客户端收到授权码，附上早先的"重定向URI"，向认证服务器申请令牌。这一步是在客户端的后台的服务器上完成的，对用户不可见。

（E）认证服务器核对了授权码和重定向URI，确认无误后，向客户端发送访问令牌（access token）和更新令牌（refresh token）。



1. 客户端携带 client_id, scope, redirect_uri, state 等信息引导用户请求授权服务器的授权端点下发 code
2. 授权服务器验证客户端身份，验证通过则询问用户是否同意授权（此时会跳转到用户能够直观看到的授权页面，等待用户点击确认授权）
3. 假设用户同意授权，此时授权服务器会将 code 和 state（如果客户端传递了该参数）拼接在 redirect_uri 后面，以302形式下发 code
4. 客户端携带 code, redirect_uri, 以及 client_secret 请求授权服务器的令牌端点下发 access_token （这一步实际上中间经过了客户端的服务器，除了 code，其它参数都是在应用服务器端添加，下文会细讲）
5. 授权服务器验证客户端身份，同时验证 code，以及 redirect_uri 是否与请求 code 时相同，验证通过后下发 access_token，并选择性下发 refresh_token
   简化模式（implicit）
   
```
   +----------+
   | Resource |
   |  Owner   |
   |          |
   +----------+
   ^
   |
   (B)
   +----|-----+          Client Identifier     +---------------+
   |         -+----(A)-- & Redirection URI --->|               |
   |  User-   |                                | Authorization |
   |  Agent  -|----(B)-- User authenticates -->|     Server    |
   |          |                                |               |
   |          |<---(C)--- Redirection URI ----<|               |
   |          |          with Access Token     +---------------+
   |          |            in Fragment
   |          |                                +---------------+
   |          |----(D)--- Redirection URI ---->|   Web-Hosted  |
   |          |          without Fragment      |     Client    |
   |          |                                |    Resource   |
   |     (F)  |<---(E)------- Script ---------<|               |
   |          |                                +---------------+
   +-|--------+
   |    |
   (A)  (G) Access Token
   |    |
   ^    v
   +---------+
   |         |
   |  Client |
   |         |
   +---------+

   Note: The lines illustrating steps (A) and (B) are broken into two
   parts as they pass through the user-agent.
```
（A）客户端将用户导向认证服务器。

（B）用户决定是否给于客户端授权。

（C）假设用户给予授权，认证服务器将用户导向客户端指定的"重定向URI"，并在URI的Hash部分包含了访问令牌。

（D）浏览器向资源服务器发出请求，其中不包括上一步收到的Hash值。

（E）资源服务器返回一个网页，其中包含的代码可以获取Hash值中的令牌。

（F）浏览器执行上一步获得的脚本，提取出令牌。

（G）浏览器将令牌发给客户端。

简易模式一般用在纯前端应用，token等信息必须保存在前端，所以此模式下token的过期时间一般设置的较短，来保证资源的一个安全性。

### 密码模式（resource owner password credentials）
```
+----------+
| Resource |
|  Owner   |
|          |
+----------+
v
|    Resource Owner
(A) Password Credentials
|
v
+---------+                                  +---------------+
|         |>--(B)---- Resource Owner ------->|               |
|         |         Password Credentials     | Authorization |
| Client  |                                  |     Server    |
|         |<--(C)---- Access Token ---------<|               |
|         |    (w/ Optional Refresh Token)   |               |
+---------+                                  +---------------+
```
（A）用户向客户端提供用户名和密码。

（B）客户端将用户名和密码发给认证服务器，向后者请求令牌。

（C）认证服务器确认无误后，向客户端提供访问令牌。

### 客户端模式（client credentials）
```
+---------+                                  +---------------+
|         |                                  |               |
|         |>--(A)- Client Authentication --->| Authorization |
| Client  |                                  |     Server    |
|         |<--(B)---- Access Token ---------<|               |
|         |                                  |               |
+---------+                                  +---------------+
```
（A）客户端向认证服务器进行身份认证，并要求一个访问令牌。

（B）认证服务器确认无误后，向客户端提供访问令牌。

Refresh Token
refresh token是用于获取access token的凭据。refresh token由授权服务器颁发给client，用于在当前访问令牌变为无效或过期时获取新的访问令牌，或者获取具有相同或更窄范围的其他访问令牌（访问令牌可能具有更短的生命周期和权限少于资源所有者授权的权限。根据授权服务器的判断，发出刷新令牌是可选的。如果授权服务器发出刷新令牌，则在发出访问令牌时包括它（即图1中的步骤（D））。
刷新令牌是表示资源所有者授予客户端的权限的字符串。该字符串通常对客户端不透明。令牌表示用于检索授权信息的标识符。与访问令牌不同，刷新令牌仅用于授权服务器，不会发送到资源服务器。
```
+--------+                                           +---------------+
|        |--(A)------- Authorization Grant --------->|               |
|        |                                           |               |
|        |<-(B)----------- Access Token -------------|               |
|        |               & Refresh Token             |               |
|        |                                           |               |
|        |                            +----------+   |               |
|        |--(C)---- Access Token ---->|          |   |               |
|        |                            |          |   |               |
|        |<-(D)- Protected Resource --| Resource |   | Authorization |
| Client |                            |  Server  |   |     Server    |
|        |--(E)---- Access Token ---->|          |   |               |
|        |                            |          |   |               |
|        |<-(F)- Invalid Token Error -|          |   |               |
|        |                            +----------+   |               |
|        |                                           |               |
|        |--(G)----------- Refresh Token ----------->|               |
|        |                                           |               |
|        |<-(H)----------- Access Token -------------|               |
+--------+           & Optional Refresh Token        +---------------+
```
（A）客户端通过向授权服务器进行认证、发起权限授予来获取access token。

（B）授权服务器验证客户端并验证权限授予授权，如果有效，则颁发访问令牌和刷新令牌。

（C）客户端通过访问令牌向资源服务器发出受保护的资源请求。

（D）资源服务器验证访问令牌，如果有效，则为请求提供服务。

（E）重复步骤（C）和（D）直到访问令牌到期。如果客户端知道访问令牌已过期，则跳到步骤（G）;否则，它会生成另一个受保护的资源请求

（F）由于访问令牌无效，资源服务器返回无效的令牌错误。

（G）客户端通过向授权服务器进行身份验证并显示刷新令牌来请求新的访问令牌。该客户端身份验证的要求是基于客户端类型和授权服务器策略。

（H）授权服务器验证客户端并验证刷新令牌，如果有效，则发出新的访问令牌（以及可选的新刷新令牌）。
