# Gin
`import "github.com/gin-gonic/gin"`

## 路由器
```go
router := gin.Default() // with logging and recovery middleware
// or
router := gin.New() // without default middleware
...
_ = router.Run(":8080")
```
## 注册路由
```go
// static routes
router.GET("/users", func(c *gin.Context) {})

// parameter routes, 如果和静态路由存在冲突, 需要先注册静态路由
router.GET("/users/:id", func(c *gin.Context) {
  idStr := c.Param("id")
  id, err := strconv.Atoi(idStr)
})

// query parameters, GET 请求的参数不需要放在路由的路径里
// URI: /users?page=1&limit=10
router.GET("/users", func(c *gin.Context) {
  page := c.Query("page")
  limit := c.DefaultQuery("limit", "20")
})
```
## Request & Response
```go
// gin 中最常用的 tag 是 json 和 binding
// binding 可以在解析报文时, 对相应字段进行基础的校验
type User struct {
  Name  string `json:"name" binding:"required"`
  Email string `json:"email" binding:"omitempty,email"`
}

// 统一响应结构体
type Response struct {
    Success bool        `json:"success"`
    Data    interface{} `json:"data,omitempty"`
    Message string      `json:"message,omitempty"`
    Error   string      `json:"error,omitempty"`
    Code    int         `json:"code,omitempty"`
}

router.POST("/users", func(c *gin.Context) {
  var user User
  // 最常用的获取请求内容的方法是 ShouldBindJSON
  // 还有一系列的 ShouldBindXXX 函数可以获取 URI 或是 http header 中的参数
  if err := c.ShouldBindJSON(&user); err != nil {
    // gin.H 是 map[string] any
    c.JSON(400, Response{
        Success: false,
        Error:   err.Error(),
        Code:    ErrorCode,
    })
    return
  }
})
```
## Middleware
router.Group

## Test

## 校验请求
