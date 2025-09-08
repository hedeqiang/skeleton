# i18n 快速参考指南

本文档提供 i18n 多语言功能的快速参考和常用代码片段。

## 🚀 5分钟快速上手

### 1. 基础配置

```yaml
# configs/config.dev.yaml
i18n:
  default_language: "zh"
  support_languages: ["zh", "en"] 
  messages_path: "./locales"
```

### 2. 消息文件

```yaml
# locales/zh.yaml  
common:
  hello: 
    other: "你好 {{.Name}}"
  success:
    other: "操作成功"

errors:
  not_found:
    other: "资源未找到"
```

```yaml
# locales/en.yaml
common:
  hello:
    other: "Hello {{.Name}}"
  success:
    other: "Operation successful"
    
errors:
  not_found:
    other: "Resource not found"
```

### 3. 代码使用

```go
// Handler 中使用
func (h *UserHandler) GetUsers(c *gin.Context) {
    msg := h.app.I18n.T(c.Request.Context(), "common.success", nil)
    c.JSON(200, gin.H{"message": msg})
}
```

### 4. 客户端调用

```bash
# 中文请求
curl -H "Accept-Language: zh-CN" http://localhost:8080/api

# 英文请求  
curl -H "Accept-Language: en-US" http://localhost:8080/api

# 查询参数方式
curl http://localhost:8080/api?lang=en
```

## 📚 常用 API

### 基础翻译

```go
// 简单消息
message := i18n.T(ctx, "common.success", nil)

// 带参数消息
greeting := i18n.T(ctx, "common.hello", map[string]interface{}{
    "Name": "Alice",
})

// 指定语言翻译
zhMsg := i18n.TWithLang("zh", "common.welcome", nil)
enMsg := i18n.TWithLang("en", "common.welcome", nil)
```

### Context 操作

```go
// 设置语言到 context
ctx := i18n.SetLanguageToContext(context.Background(), "en")

// 获取 context 中的语言
lang := i18n.GetLanguageFromContext(ctx)

// 解析 Accept-Language
lang := i18n.ParseAcceptLanguage("zh-CN,zh;q=0.9,en;q=0.8")
```

### 错误处理

```go
// 创建 i18n 错误
err := errors.NewI18n(errors.ErrorTypeValidation, "errors.required", 
    map[string]interface{}{"Field": "username"})

// 获取本地化错误消息  
if appErr, ok := err.(*errors.AppError); ok {
    localMsg := appErr.LocalizedMessage(ctx, i18n)
}

// 使用预定义错误
return errors.ErrUserNotFound // 自动支持 i18n
```

## 🎯 常用消息模板

### 基础消息

```yaml
# 通用状态
common:
  success: 
    other: "操作成功"
  failed:
    other: "操作失败"  
  loading:
    other: "加载中..."
  welcome:
    other: "欢迎！"

# 操作确认
confirm:
  delete:
    other: "确定要删除吗？"
  save:
    other: "确定要保存吗？"
```

### 业务消息

```yaml
# 用户相关
user:
  created:
    other: "用户创建成功"
  updated: 
    other: "用户更新成功"
  deleted:
    other: "用户删除成功"
  login_success:
    other: "登录成功"
  logout_success:
    other: "退出成功"

# 订单相关  
order:
  placed:
    other: "订单已提交"
  cancelled:
    other: "订单已取消"
  shipped:
    other: "订单已发货"
```

### 错误消息

```yaml
# 通用错误
errors:
  not_found:
    other: "资源不存在"
  unauthorized:
    other: "未授权访问"
  forbidden:
    other: "访问被禁止"
  internal_error:
    other: "内部服务器错误"
  
# 验证错误
validation:
  required:
    other: "{{.Field}}是必填项"
  invalid_email:
    other: "邮箱格式不正确"
  password_too_short:
    other: "密码至少需要{{.MinLength}}个字符"
  
# 业务错误  
business:
  user_exists:
    other: "用户已存在"
  insufficient_balance:
    other: "余额不足"
  order_not_found:
    other: "订单不存在"
```

## 🔧 配置参考

### 完整配置示例

```yaml
i18n:
  default_language: "zh"                    # 默认语言
  support_languages: ["zh", "en", "ja"]    # 支持语言列表
  messages_path: "./locales"                # 消息文件目录
  accept_languages:                         # Accept-Language 识别列表
    - "zh"
    - "zh-CN" 
    - "zh-Hans"
    - "en"
    - "en-US"
    - "ja"
    - "ja-JP"
```

### 环境变量覆盖

```bash
# 设置默认语言
export I18N_DEFAULT_LANGUAGE=en

# 设置消息文件路径  
export I18N_MESSAGES_PATH=/app/locales

# 设置支持语言 (逗号分隔)
export I18N_SUPPORT_LANGUAGES=zh,en,ja
```

## 🛠️ 开发工具

### 消息完整性检查

```bash
# 检查两种语言的消息键是否一致
diff <(yq eval 'keys | .[]' locales/zh.yaml | sort) \
     <(yq eval 'keys | .[]' locales/en.yaml | sort)
```

### YAML 语法验证

```bash
# 验证 YAML 语法
yamllint locales/zh.yaml
yamllint locales/en.yaml

# 或使用 yq
yq eval '.' locales/zh.yaml > /dev/null && echo "zh.yaml OK"
yq eval '.' locales/en.yaml > /dev/null && echo "en.yaml OK"
```

### 消息键统计

```bash
# 统计消息数量
echo "中文消息数: $(yq eval '[.. | select(type == "string" and . != "")] | length' locales/zh.yaml)"
echo "英文消息数: $(yq eval '[.. | select(type == "string" and . != "")] | length' locales/en.yaml)"
```

## 🧪 测试代码片段

### 单元测试

```go
func TestI18nTranslation(t *testing.T) {
    i18n, err := i18n.New(i18n.Config{
        DefaultLanguage: "zh",
        SupportLangs:    []string{"zh", "en"},
        MessagesPath:    "./locales",
    })
    require.NoError(t, err)
    
    // 测试中文翻译
    ctx := i18n.SetLanguageToContext(context.Background(), "zh")
    msg := i18n.T(ctx, "common.success", nil)
    assert.Equal(t, "操作成功", msg)
    
    // 测试英文翻译
    ctx = i18n.SetLanguageToContext(context.Background(), "en")
    msg = i18n.T(ctx, "common.success", nil)
    assert.Equal(t, "Operation successful", msg)
    
    // 测试带参数翻译
    ctx = i18n.SetLanguageToContext(context.Background(), "zh")
    msg = i18n.T(ctx, "common.hello", map[string]interface{}{"Name": "张三"})
    assert.Equal(t, "你好 张三", msg)
}
```

### 集成测试

```go
func TestI18nMiddleware(t *testing.T) {
    // 设置测试环境
    gin.SetMode(gin.TestMode)
    r := gin.New()
    
    // 创建 i18n 实例
    i18n, err := i18n.New(i18n.Config{...})
    require.NoError(t, err)
    
    // 注册中间件
    r.Use(middleware.NewI18n(i18n))
    
    // 注册测试路由
    r.GET("/test", func(c *gin.Context) {
        msg := i18n.T(c.Request.Context(), "common.success", nil)
        c.JSON(200, gin.H{"message": msg})
    })
    
    // 测试中文请求
    req := httptest.NewRequest("GET", "/test", nil)
    req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")
    resp := httptest.NewRecorder()
    r.ServeHTTP(resp, req)
    
    assert.Equal(t, 200, resp.Code)
    assert.Contains(t, resp.Body.String(), "操作成功")
    assert.Equal(t, "zh", resp.Header().Get("Content-Language"))
}
```

## 🎨 Handler 模板

### 标准 CRUD 操作

```go
// 创建资源
func (h *UserHandler) CreateUser(c *gin.Context) {
    // ... 业务逻辑
    
    if err != nil {
        // 返回本地化错误
        if appErr, ok := err.(*errors.AppError); ok {
            localMsg := appErr.LocalizedMessage(c.Request.Context(), h.app.I18n)
            c.JSON(appErr.StatusCode(), gin.H{"error": localMsg})
            return
        }
        
        // 通用错误处理
        msg := h.app.I18n.T(c.Request.Context(), "errors.internal_error", nil)
        c.JSON(500, gin.H{"error": msg})
        return
    }
    
    // 返回成功消息
    successMsg := h.app.I18n.T(c.Request.Context(), "user.created", nil)
    c.JSON(201, gin.H{
        "message": successMsg,
        "data":    user,
    })
}

// 获取资源列表
func (h *UserHandler) ListUsers(c *gin.Context) {
    users, err := h.userService.ListUsers(c.Request.Context())
    if err != nil {
        msg := h.app.I18n.T(c.Request.Context(), "errors.internal_error", nil)
        c.JSON(500, gin.H{"error": msg})
        return
    }
    
    c.JSON(200, gin.H{
        "data":  users,
        "count": len(users),
    })
}
```

### 验证错误处理

```go
func (h *UserHandler) CreateUser(c *gin.Context) {
    var req CreateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        // 处理验证错误
        if validationErrs, ok := err.(validator.ValidationErrors); ok {
            errorMsgs := make([]string, 0)
            for _, validationErr := range validationErrs {
                field := validationErr.Field()
                tag := validationErr.Tag()
                
                var msgKey string
                var data map[string]interface{}
                
                switch tag {
                case "required":
                    msgKey = "validation.required"
                    data = map[string]interface{}{"Field": field}
                case "min":
                    msgKey = "validation.min_length"
                    data = map[string]interface{}{
                        "Field": field,
                        "Min":   validationErr.Param(),
                    }
                default:
                    msgKey = "validation.invalid"
                    data = map[string]interface{}{"Field": field}
                }
                
                localMsg := h.app.I18n.T(c.Request.Context(), msgKey, data)
                errorMsgs = append(errorMsgs, localMsg)
            }
            
            c.JSON(400, gin.H{"errors": errorMsgs})
            return
        }
        
        // 通用绑定错误
        msg := h.app.I18n.T(c.Request.Context(), "errors.invalid_input", nil)
        c.JSON(400, gin.H{"error": msg})
        return
    }
    
    // 继续处理...
}
```

## 🔍 调试技巧

### 启用调试日志

```go
// 在 i18n.T 方法中添加调试日志
func (i *I18n) T(ctx context.Context, messageID string, templateData map[string]interface{}) string {
    lang := GetLanguageFromContext(ctx)
    if lang == "" {
        lang = i.defaultLang
    }
    
    // 调试日志
    log.Printf("[i18n] Translating '%s' to '%s' with data: %+v", messageID, lang, templateData)
    
    // ... 原有逻辑
}
```

### HTTP 请求调试

```bash
# 查看详细的 HTTP 交互
curl -v -H "Accept-Language: zh-CN,zh;q=0.9" http://localhost:8080/api/users

# 检查响应头中的 Content-Language
curl -I -H "Accept-Language: en-US" http://localhost:8080/api/users
```

### 消息查找调试

```bash
# 查找特定消息键
grep -r "user_not_found" locales/

# 查找包含特定文本的消息
grep -r "操作成功" locales/

# 列出所有消息键
yq eval 'paths(scalar) as $p | $p | join(".")' locales/zh.yaml
```

---

## 📞 获取帮助

- 📖 完整文档: [I18N.md](./I18N.md)
- 🐛 问题反馈: 项目 Issues
- 💡 功能建议: 项目 Discussions

---

*这份快速参考指南涵盖了 i18n 功能的 80% 常用场景，更多高级用法请参考完整文档。*